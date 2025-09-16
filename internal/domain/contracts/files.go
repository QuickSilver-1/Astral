package contracts

import (
	"astral/internal/domain/file"
)

type FilesInterface interface {
	UploadFiles(fileData file.File) (*file.File, error)
	GetFilesByUser(userID string, filter FilterData) ([]file.File, error)
	GetFileByID(ID, userID string) (*file.File, error)
	DeleteFile(ID, userID string) (*file.File, error)
}

type FilterData struct {
	Key   string
	Value string
	Limit int
}

func NewFilterData(value, key string, limit int) *FilterData {
	return &FilterData{
		Value: value,
		Key:   key,
		Limit: limit,
	}
}