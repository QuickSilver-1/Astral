package fileservice

import (
	"astral/internal/domain/contracts"
	"astral/internal/domain/file"
	"astral/internal/repository/db/redis"
	filesrepo "astral/internal/repository/files"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	FILE_LOAD_TIMEOUT = time.Second*60
	DEFAULT_TIMEOUT   = time.Second*5
)

type FilesService struct {
	repo 	filesrepo.StorageRepo
	cash    redis.CashStorage
	logger 	*slog.Logger
}

func NewFileService(repo filesrepo.StorageRepo, cash redis.CashStorage, logger *slog.Logger) *FilesService {
	return &FilesService{
		repo: 	repo,
		cash: 	cash,
		logger: logger,
	}
}

func (s *FilesService) UploadFiles(fileData file.File) (*file.File, error) {
	const op = "service.files.UploadFiles"
	s.logger.Info("Usecase start", "func", op, "filename", fileData.Name, "userID", fileData.User)

	ctx, cancel := context.WithTimeout(context.Background(), FILE_LOAD_TIMEOUT)
	defer cancel()
	out := make(chan error)
	go s.cash.CashedFile(ctx, generateKeyForCash("filename:", fileData.Name), fileData, out)

	ctx, cancel = context.WithTimeout(context.Background(), FILE_LOAD_TIMEOUT)
	defer cancel()

	res, err := s.repo.CreateFile(ctx, fileData.User, fileData)
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()
	s.cash.DelKeyByPrefix(ctx, generateKeyForCash("list:" + fileData.User, ""))

	<-out

	return res, nil
}

func (s *FilesService) GetFilesByUser(userID string, filter contracts.FilterData) ([]file.File, error) {
	const op = "service.files.GetFilesForUser"
	s.logger.Info("Usecase start", "func", op, "userID", userID)

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	jsonData := s.cash.GetKey(ctx, generateKeyForCash("list:" + userID, filter))
	if jsonData != "" {
		var files []file.File
		err := json.Unmarshal([]byte(jsonData), &files)
		if err != nil {
			s.logger.Warn("failed to unmarshal data", "func", op, "data", jsonData, "error", err)
		}

		return files, nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	files, err := s.repo.ListUserFiles(ctx, userID)
	if err != nil {
		return nil, err
	}

	filteredFiles := s.applyFilters(files, filter)

    if filter.Limit > 0 && len(filteredFiles) > filter.Limit {
        filteredFiles = filteredFiles[:filter.Limit]
    }

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()
	s.cash.NewKey(ctx, generateKeyForCash("list:" + userID, filter), filteredFiles)

	return filteredFiles, nil
}

func (s *FilesService) GetFileByID(ID, userID string) (*file.File, error) {
	const op = "service.files.GetFileByID"
	s.logger.Info("Usecase start", "func", op)

	ctx, cancel := context.WithTimeout(context.Background(), FILE_LOAD_TIMEOUT)
	defer cancel()

	fileData := s.cash.GetCashedFile(ctx, generateKeyForCash("file:", ID))
	if fileData != nil {
		return fileData, nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	fileData, err := s.repo.GetFileInfo(ctx, userID, ID)
	if err != nil {
		return nil, err
	}

	if fileData.Metadata["public"] == "false" && fileData.User != userID && slices.Contains[[]string, string](fileData.Grant, userID) {
		s.logger.Info("access denied", "func", op, "fileID", ID, "userID", userID)
		return nil, ErrAccessDenied
	}

	ctx, cancel = context.WithTimeout(context.Background(), FILE_LOAD_TIMEOUT)
	defer cancel()

	reader, err := s.repo.GetFileByID(ctx, userID, ID)
	if err != nil {
		return nil, err
	}
	fileData.Reader = reader

	ctx, cancel = context.WithTimeout(context.Background(), FILE_LOAD_TIMEOUT)
	defer cancel()
	out := make(chan error)
	go s.cash.CashedFile(ctx, generateKeyForCash("file:", ID), *fileData, out)

	return fileData, nil
}

func (s *FilesService) DeleteFile(ID, userID string) (*file.File, error) {
	const op = "service.files.DeleteFile"
	s.logger.Info("Usecase start", "func", op, "fileID", ID, "userID", userID)

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	fileInfo, err := s.repo.GetFileInfo(ctx, userID, ID)
	if err != nil {
		return nil, err
	}

	if fileInfo.User != userID {
		s.logger.Info("access denied", "func", op, "fileID", ID, "userID", userID)
		return nil, ErrAccessDenied
	}

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	err = s.repo.DeleteFile(ctx, ID, userID)
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()
	go s.cash.DelKey(ctx, generateKeyForCash("file:", ID))

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()
	s.cash.DelKeyByPrefix(ctx, generateKeyForCash("list:" + userID, ""))

	return fileInfo, nil
}

func (s *FilesService) applyFilters(files []file.File, filter contracts.FilterData) []file.File {
    if filter.Key == "" || filter.Value == "" {
        return files
    }

    var filtered []file.File

    for _, file := range files {
        if s.matchesFilter(file, filter) {
            filtered = append(filtered, file)
        }
    }

    return filtered
}

func (s *FilesService) matchesFilter(file file.File, filter contracts.FilterData) bool {
    switch filter.Key {
    case "name":
        return strings.Contains(strings.ToLower(file.Name), strings.ToLower(filter.Value))
    
    case "mime":
        return strings.Contains(strings.ToLower(file.Mime), strings.ToLower(filter.Value))
    
    case "public":
        switch filter.Value {
		case "true":
            return file.Public
        case "false":
            return !file.Public
        }
        return false
    
    case "file":
        switch filter.Value {
		case "true":
            return file.File
        case "false":
            return !file.File
        }
        return false
    
    case "size":
        return s.filterBySize(file, filter.Value)
    
    case "created":
        return s.filterByDate(file, filter.Value)
    
    case "metadata":
        return s.filterByMetadata(file, filter.Value)
    
    case "grant":
        return s.filterByGrant(file, filter.Value)
    
    default:
        return false
    }
}

func (s *FilesService) filterBySize(file file.File, value string) bool {
    if strings.Contains(value, ">") {
        minSize, err := strconv.Atoi(strings.TrimPrefix(value, ">"))
        return err == nil && file.Size > minSize
    }
    
    if strings.Contains(value, "<") {
        maxSize, err := strconv.Atoi(strings.TrimPrefix(value, "<"))
        return err == nil && file.Size < maxSize
    }
    
    if strings.Contains(value, "-") {
        parts := strings.Split(value, "-")
        if len(parts) == 2 {
            min, err1 := strconv.Atoi(parts[0])
            max, err2 := strconv.Atoi(parts[1])
            return err1 == nil && err2 == nil && file.Size >= min && file.Size <= max
        }
    }
    
    size, err := strconv.Atoi(value)
    return err == nil && file.Size == size
}

func (s *FilesService) filterByDate(file file.File, value string) bool {
    if file.CreatedAt == nil {
        return false
    }
    
    now := time.Now()
    
    switch value {
    case "today":
        return file.CreatedAt.Year() == now.Year() &&
               file.CreatedAt.Month() == now.Month() &&
               file.CreatedAt.Day() == now.Day()
    
    case "week":
        weekAgo := now.AddDate(0, 0, -7)
        return file.CreatedAt.After(weekAgo)
    
    case "month":
        monthAgo := now.AddDate(0, -1, 0)
        return file.CreatedAt.After(monthAgo)
    
    default:
        filterDate, err := time.Parse("2006-01-02", value)
        if err == nil {
            return file.CreatedAt.Year() == filterDate.Year() &&
                   file.CreatedAt.Month() == filterDate.Month() &&
                   file.CreatedAt.Day() == filterDate.Day()
        }
        return false
    }
}

func (s *FilesService) filterByMetadata(file file.File, value string) bool {
    for _, v := range file.Metadata {
        if strings.Contains(strings.ToLower(v), strings.ToLower(value)) {
            return true
        }
    }
    return false
}

func (s *FilesService) filterByGrant(file file.File, value string) bool {
    for _, grant := range file.Grant {
        if strings.EqualFold(grant, value) {
            return true
        }
    }
    return false
}

func generateKeyForCash(query string, data any) string {
	dataByte, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	dataStr := string(dataByte)
	return fmt.Sprintf("%s:%s", query, dataStr)
}