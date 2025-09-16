package miniostorage

import (
	"context"
	"fmt"
	"time"

	"astral/env"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	Client     *minio.Client
	BucketName string
}

func NewMinioStorage(cfg *env.MinIO) (*MinioStorage, error) {
	const op = "storage.minio.New"

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create MinIO client: %w", op, err)
	}

	mStorage := &MinioStorage{
		Client:     client,
		BucketName: cfg.BucketName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := mStorage.ensureBucketExists(ctx, cfg.BucketName); err != nil {
		return nil, fmt.Errorf("%s: filed to ensure bucket exists: %w", op, err)
	}

	return mStorage, nil
}

func (s *MinioStorage) ensureBucketExists(ctx context.Context, bucketName string) error {
	const op = "storage.minio.ensureBucketExists"

	exists, err := s.bucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("%s: failed to check bucket existence: %w", op, err)
	}

	if !exists {
		if err := s.makeBucket(ctx, bucketName); err != nil {
			return fmt.Errorf("%s: failed to create bucket: %w", op, err)
		}
	}

	return nil
}

func (s *MinioStorage) bucketExists(ctx context.Context, bucketName string) (bool, error) {
	const op = "storage.minio.bucketExists"

	exists, err := s.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("%s: failed to check bucket existence: %w", op, err)
	}

	return exists, nil
}

func (s *MinioStorage) makeBucket(ctx context.Context, bucketName string) error {
	const op = "storage.minio.makeBucket"

	exists, err := s.bucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("%s: failed to check bucket existence: %w", op, err)
	}

	if exists {
		return nil
	}

	err = s.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("%s: failed to create bucket: %w", op, err)
	}

	return nil
}
