package filesrepo

import (
	"astral/internal/domain/file"
	miniostorage "astral/internal/repository/db/minio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	MAX_FILE_SIZE = 100 * 1024 * 1024
)

type StoragePersister struct {
	logger *slog.Logger
	storage miniostorage.MinioStorage
}

func NewFilePersister(storage miniostorage.MinioStorage, logger *slog.Logger) *StoragePersister {
	return &StoragePersister{
		storage: storage,
		logger:  logger,
	}
}

func (s *StoragePersister) CreateFile(ctx context.Context, userID string, fileData file.File) (*file.File, error) {
    const op = "storage.minio.UploadProjectFile"

    if err := validateFileName(fileData.Name); err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    if fileData.Size > MAX_FILE_SIZE {
		s.logger.Info("file size exceeds limit", "func", op, "filename", fileData.Name, "userID", userID, "size", fileData.Size)
        return nil, NewErrFileUpload("file size exceeds limit")
    }

    contentType := detectContentType(fileData.Name)

	fileID := uuid.New().String()
    filePath := getFilePath(userID, fileID)

    safeMetadata := make(map[string]string)
    for k, v := range fileData.Metadata {
        safeKey := strings.TrimSpace(k)
        safeValue := strings.TrimSpace(v)
        if safeKey != "" && safeValue != "" {
            safeMetadata[safeKey] = safeValue
        }
    }

	safeMetadata["file_name"] = fileData.Name
	safeMetadata["grant"] = strings.Join(fileData.Grant, ";")
	if fileData.Public {
		safeMetadata["public"] = "true"
	} else {
		safeMetadata["public"] = "false"
	}

    putOptions := minio.PutObjectOptions{
        ContentType:  contentType,
        UserMetadata: safeMetadata,
    }

    info, err := s.storage.Client.PutObject(ctx, s.storage.BucketName, filePath, fileData.Reader, int64(fileData.Size), putOptions)
    if err != nil {
		s.logger.Error("failed to upload file in minio", "func", op, "filename", fileData.Name, "userID", userID, "error", err)
        return nil, errors.New("failed to upload file")
    }

    if info.Size == 0 {
		s.logger.Error("failed to upload file", "func", op, "filename", fileData.Name, "userID", userID, "error", err)
		return nil, errors.New("failed to upload file")
    }

    result := &file.File{
		ID: 	   fileID,
        Name:      fileData.Name,
        Public:    fileData.Public,
        Mime:      fileData.Mime,
		File:      fileData.File,
        Grant:     fileData.Grant,
		Metadata:  fileData.Metadata,
		CreatedAt: &info.LastModified,
    }

    return result, nil
}

func validateFileName(fileName string) error {
    if fileName == "" {
        return NewErrFileUpload("empty file name")
    }
    if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") {
        return NewErrFileUpload("invalid file name")
    }
    if len(fileName) > 255 {
        return NewErrFileUpload("file name is too long")
    }
    return nil
}

func (s *StoragePersister) GetFileByID(ctx context.Context, userID, fileID string) (io.ReadCloser, error) {
	filePath := getFilePath(userID, fileID)
	return s.download(ctx, filePath)
}

func (s *StoragePersister) DeleteFile(ctx context.Context, fileID, userID string) error {
	return s.delete(ctx, fileID, userID)
}

func (s *StoragePersister) GetFileInfo(ctx context.Context, userID, fileID string) (*file.File, error) {
	filePath := getFilePath(userID, fileID)
	return s.getFileInfo(ctx, filePath)
}

func (s *StoragePersister) ListUserFiles(ctx context.Context, userID string) ([]file.File, error) {
	return s.listFiles(ctx, userID+"/")
}

func (s *StoragePersister) listFiles(ctx context.Context, prefix string) ([]file.File, error) {
	const op = "storage.minio.listFiles"

	objectCh := s.storage.Client.ListObjects(ctx, s.storage.BucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var files []file.File

	for objInfo := range objectCh {
		if objInfo.Err != nil {
			s.logger.Error("error listing objects", "func", op, "error", objInfo.Err)
			return nil, errors.New("error listing files")
		}

		grant := strings.Split(objInfo.UserMetadata["grant"], ";")
		files = append(files, file.File{
			ID:           strings.Join(strings.Split(objInfo.Key, "/")[1:], ""),
			Name:         objInfo.UserMetadata["file_name"],
			File:         int(objInfo.Size) != 0,
			Public:   	  objInfo.UserMetadata["public"] == "true",
			Mime:  		  objInfo.ContentType,
			Grant:        grant,
			Size:         int(objInfo.Size),
			CreatedAt:    &objInfo.LastModified,
			Metadata:     objInfo.UserMetadata,
		})
	}

	return files, nil
}

func detectContentType(fileName string) string {
	switch path.Ext(fileName) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".zip":
		return "application/zip"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}

func (s *StoragePersister) download(ctx context.Context, path string) (io.ReadCloser, error) {
	const op = "storage.minio.download"

	obj, err := s.storage.Client.GetObject(ctx, s.storage.BucketName, path, minio.GetObjectOptions{})
	if err != nil {
		s.logger.Error("failed to get file from minio", "func", op, "error", err)
		return nil, errors.New("failed to get file from minio")
	}

	_, err = obj.Stat()
	if err != nil {
		if err.(minio.ErrorResponse).Code == "NoSuchKey" {
			return nil, ErrFileNotFound
		}

		s.logger.Error("failed to get file info", "func", op, "path", path, "error", err)
		return nil, errors.New("failed to get file info")
	}

	return obj, nil
}

func (s *StoragePersister) delete(ctx context.Context, fileID, userID string) error {
	const op = "storage.minio.delete"

	path := getFilePath(userID, fileID)
	err := s.storage.Client.RemoveObject(ctx, s.storage.BucketName, path, minio.RemoveObjectOptions{})
	if err != nil {
		if err.(minio.ErrorResponse).Code == "NoSuchKey" {
			return ErrFileNotFound
		}

		s.logger.Error("failed to delete file info", "func", op, "path", path, "error", err)
		return errors.New("failed to delete file info")
	}

	return nil
}

func (s *StoragePersister) getFileInfo(ctx context.Context, path string) (*file.File, error) {
	const op = "storage.minio.getFileInfo"

	objInfo, err := s.storage.Client.StatObject(ctx, s.storage.BucketName, path, minio.StatObjectOptions{})
	if err != nil {
		if err.(minio.ErrorResponse).Code == "NoSuchKey" {
			return nil, ErrFileNotFound
		}

		s.logger.Error("failed to get file info", "func", op, "path", path, "error", err)
		return nil, errors.New("failed to get file info")
	}

	grant := strings.Split(objInfo.UserMetadata["grant"], ";")
	delete(objInfo.UserMetadata, "grant")
	delete(objInfo.UserMetadata, "file_name")
	delete(objInfo.UserMetadata, "public")

	return &file.File{
			ID:           strings.Join(strings.Split(objInfo.Key, "/")[1:], ""),
			Name:         objInfo.UserMetadata["file_name"],
			File:         int(objInfo.Size) != 0,
			Public:   	  objInfo.UserMetadata["public"] == "true",
			Mime:  		  objInfo.ContentType,
			Grant:        grant,
			Size:         int(objInfo.Size),
			CreatedAt:    &objInfo.LastModified,
			Metadata:     objInfo.UserMetadata,
			User:         strings.Split(objInfo.Key, "/")[0],
		}, nil
}
