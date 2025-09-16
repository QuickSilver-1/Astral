package main

import (
	"astral/env"
	"astral/internal/presentation"
	authrepo "astral/internal/repository/auth"
	miniostorage "astral/internal/repository/db/minio"
	pg "astral/internal/repository/db/postgres"
	"astral/internal/repository/db/redis"
	filesrepo "astral/internal/repository/files"
	authservice "astral/internal/services/authorization"
	fileservice "astral/internal/services/files"
	validationservice "astral/internal/services/validation"
	"astral/logger"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	env := env.MustLoad()
	logger := logger.NewLogger(env.Env)
	
	pgStorage, err := pg.NewDBConnection(&env.PgSql)
	if err != nil {
		logger.Error("failed to connect to database postgres", "error", err, "host", env.PgSql.Host)
		return
	}

	validatonService := validationservice.NewValidationService(logger)

	authPersister := authrepo.NewUserPersister(pgStorage, logger)
	authService := authservice.NewAuthService(authPersister, logger, validatonService, env.JWTSecret, env.AdminToken)

	minioStorage, err := miniostorage.NewMinioStorage(&env.MinIO)
	if err != nil {
		logger.Error("failed to connect to database minio", "error", err, "bucket", env.MinIO.BucketName)
		return
	}

	filesPersister := filesrepo.NewFilePersister(*minioStorage, logger)
	cachPersister, err := redis.NewConnectRedis(env.Redis, logger)
	if err != nil {
		logger.Error("failed to connect to database redis", "error", err, "bucket", env.MinIO.BucketName)
		return
	}
	fileService := fileservice.NewFileService(filesPersister, *cachPersister, logger)

	app := presentation.New(logger, env, authService, validatonService, fileService)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := app.Run(); err != nil {
			logger.Error("server error", slog.String("error", err.Error()))
			quit <- syscall.SIGTERM
		}
	}()

	logger.Info("Application started successfully")

	<-quit
	logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var errors []error
	var mu sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()

		ctx = context.Background()
		if err := app.GracefulShutdown(ctx); err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("server shutdown error: %w", err))
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
	}()

	if len(errors) > 0 {
		logger.Info("Application has been shutdown with errors", "errors", errors)
	} else {
		logger.Info("Application has been shutdown gracefully")
	}
}