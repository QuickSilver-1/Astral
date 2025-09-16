package presentation

import (
	"astral/env"
	"astral/internal/domain/contracts"
	server "astral/internal/presentation/http"
	"context"
	"log/slog"
)

const (
	EXIT_CHAN_KEY ctxKey = "chan"
)

type ctxKey string

type Api struct {
	logger   		  *slog.Logger
	server 			  *server.Server
	port   			  string
}

func New(
	logger 			  *slog.Logger,
	env 			  *env.Env,
	authService 	  contracts.AuthInterface,
	validationService contracts.ValidationInterface,
	filesService      contracts.FilesInterface,
) *Api {
	port := env.Http.GetPort()
	router := server.NewHandler(logger, filesService, authService, validationService, *env)
	routes := router.InitRouts(env)
	srv := server.NewServer(port, routes)

	return &Api{
		logger: logger,
		server: srv,
		port:   port,
	}
}

func (a *Api) Run() error {
	a.logger.Info("Starting the application", slog.String("port", a.port))

	return a.server.Run()
}

func (a *Api) GracefulShutdown(ctx context.Context) error {
	a.logger.Info("Initiating graceful shutdown")

	return a.server.Shutdown(ctx)
}
