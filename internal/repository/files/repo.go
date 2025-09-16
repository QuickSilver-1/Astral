package filesrepo

import (
	"astral/internal/domain/file"
	"context"
	"io"
)

type StorageRepo interface {
	CreateFile(ctx context.Context, userID string, fileData file.File) (*file.File, error)
	GetFileByID(ctx context.Context, userID, fileID string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, fileID, userID string) error
	GetFileInfo(ctx context.Context, userID, fileID string) (*file.File, error)
	ListUserFiles(ctx context.Context, userID string) ([]file.File, error)
}
