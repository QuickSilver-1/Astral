package redis

import (
	"astral/env"
	"astral/internal/domain/file"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
)

type CashStorage struct {
	DB     *redis.Client
	logger *slog.Logger
}

func NewConnectRedis(cfg env.Redis, logger *slog.Logger) (*CashStorage, error) {
	op := "db.redis.CashStorage.NewConnectRedis"

    r := redis.NewClient(&redis.Options{
        Addr:     cfg.Addr,
        Password: cfg.Pass,
        DB:       0,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err := r.Ping(ctx).Result()
    if err != nil {
		logger.Error("failed to connect to Redis", "func", op, "address", cfg.Addr, "password", cfg.Pass, "error", err)
        return nil, errors.New("failed to connect to Redis")
    }

    logger.Info("Redis connection established successfully")
    
    return &CashStorage{
        DB:     r,
        logger: logger,
    }, nil
}

func (s *CashStorage) NewKey(ctx context.Context, key string, value any) error {
	jsonData, err := json.Marshal(value)
    if err != nil {
		s.logger.Error("faild to marshal data", "key", key, "value", value, "error", err)
		return errors.New("failed to marshal data")
    }

	ans := s.DB.Set(ctx, key, jsonData, time.Minute*15)
    _, err = ans.Result()

    if err != nil {
        s.logger.Error("faild to create redis key", "key", key, "value", value, "error", err)
        return errors.New("faild to create redis key")
    }

    return nil
}

func (s *CashStorage) CashedFile(ctx context.Context, key string, value file.File, out chan error) {
	op := "db.redis.CashStorage.CashedFile"

	data, err := io.ReadAll(value.Reader)
	if err != nil {
		s.logger.Error("failed to read file", "func", op, "key", key, "filename", value.Name, "error", err)
		out<- nil
        return
	}

	fileData := CashedFile{
		ID: value.ID,
		Name: value.Name,
		File: value.File,
		Public: value.Public,
		Mime: value.Mime,
		Grant: value.Grant,
		Size: value.Size,
		Metadata: value.Metadata,
		CreatedAt: value.CreatedAt,
		User: value.User,
		Data: data,
	}

	jsonData, err := json.Marshal(fileData)
    if err != nil {
		s.logger.Error("faild to marshal data", "key", key, "value", value, "error", err)
		out<- errors.New("failed to marshal data")
        return
    }

	ans := s.DB.Set(ctx, key, jsonData, time.Minute*15)
    _, err = ans.Result()
    if err != nil {
        s.logger.Error("faild to create redis key", "key", key, "value", value, "error", err)
        out<- errors.New("faild to create redis key")
        return
    }

    out<- nil
}

func (s *CashStorage) GetCashedFile(ctx context.Context, key string) *file.File {
    op := "db.redis.CashStorage.GetCashedFile"

    code := s.DB.Get(ctx, key)
    value := code.Val()

    var cachedFile CashedFile
    err := json.Unmarshal([]byte(value), &cachedFile)
    if err != nil {
        s.logger.Warn("filed to unmarshal data", "func", op, "key", key, "error", err)
        return nil
    }

    return &file.File{
        ID:         cachedFile.ID,
        Name:       cachedFile.Name,
        File:       cachedFile.File,
        Public:     cachedFile.Public,
        Mime:       cachedFile.Mime,
        Grant:      cachedFile.Grant,
        Size:       cachedFile.Size,
        Metadata:   cachedFile.Metadata,
        CreatedAt:  cachedFile.CreatedAt,
        User:       cachedFile.User,
        Reader:     bytes.NewReader(cachedFile.Data),
    }
}

func (s *CashStorage) GetKey(ctx context.Context, key string) string {
    code := s.DB.Get(ctx, key)
    return code.Val()
}

func (s *CashStorage) DelKey(ctx context.Context, key string) error {
	op := "db.redis.CashStorage.DelKey"

    err := s.DB.Del(ctx, key).Err()
    if err != nil {
        s.logger.Error("failed to delete key", "func", op, "key", key, "error", err)
        return errors.New("failed to delete key")
    }

    return nil
}

func (s *CashStorage) DelKeyByPrefix(ctx context.Context, prefix string) error {
	op := "db.redis.CashStorage.DelKeyByPrefix"
	var cursor uint64
	const batchSize = 5000
	const maxIterations = 1000
	iteration := 0

	s.logger.Info("starting prefix deletion", "func", op, "prefix", prefix)

	for {
		if err := ctx.Err(); err != nil {
			s.logger.Warn("operation cancelled", "func", op, "prefix", prefix, "error", err)
			return err
		}

		keys, nextCursor, err := s.DB.Scan(ctx, cursor, prefix+"*", batchSize).Result()
		if err != nil {
			s.logger.Error("scan failed", "func", op, "prefix", prefix, "error", err)
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		if len(keys) > 0 {
			pipe := s.DB.Pipeline()
			for _, key := range keys {
				pipe.Del(ctx, key)
			}
			_, err := pipe.Exec(ctx)
			if err != nil {
				s.logger.Error("pipeline delete failed", "func", op, "prefix", prefix, "error", err, "keys_count", len(keys))
				return fmt.Errorf("failed to delete keys with pipeline: %w", err)
			}
			s.logger.Debug("batch deleted", "func", op, "batch_size", len(keys))
		}

		cursor = nextCursor
		iteration++

		if iteration > maxIterations {
			s.logger.Error("too many iterations", "func", op, "prefix", prefix, "iteration", iteration)
			return errors.New("del key by prefix exceeded max iterations")
		}

		if cursor == 0 {
			break
		}
	}

	s.logger.Info("prefix deletion completed", "func", op, "prefix", prefix)
	return nil
}
