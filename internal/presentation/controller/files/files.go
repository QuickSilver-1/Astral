package filescontroller

import (
	"astral/internal/domain/contracts"
	"astral/internal/domain/file"
	controllererrors "astral/internal/presentation/controller/errors"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// @Summary Upload document
// @Description Upload new document with metadata and file
// @Tags docs
// @Accept multipart/form-data
// @Produce json
// @Param meta formData string true "Document metadata in JSON format" example({"name": "photo.jpg", "file": true, "public": false, "mime": "image/jpg", "grant": ["login1", "login2"]})
// @Param json formData string false "Document data in JSON format (optional)"
// @Param file formData file true "Document file"
// @Success 200 {object} uploadDataResponse "Document uploaded successfully"
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 409 {object} response.ErrorResponse "Conflict"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/docs [post]
func (c *Controller) UploadFile(ctx *gin.Context) {
	token := c.utils.GetTokenFromHeader(ctx)
	if token == nil {
		return
	}

	metaJSON := ctx.PostForm("meta")
    if metaJSON == "" {
        c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("meta field is required"))
        return
    }

	var meta docMeta
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
        c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("invalid meta"))
        return
    }

    if meta.Name == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("name field is required"))
        return
    }

	var documentData map[string]any
    if jsonData := ctx.PostForm("json"); jsonData != "" {
        if err := json.Unmarshal([]byte(jsonData), &documentData); err != nil {
            c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("invalid json"))
            return
        }
    }

	form, err := ctx.MultipartForm()
	if err != nil {
		c.logger.Error("failed to get multipart form", "err", err)
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("invalid multipart form"))
		return
	}

	files := form.File["file"]
	if len(files) != 1 {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("cannot load more than 1 file for one time"))
		return
	}

	r, err := files[0].Open()
	if err != nil {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("failed to open file"))
	}

	if meta.Mime == "" {
		fileType := files[0].Header.Get("Content-Type")
		meta.Mime = fileType
	}
	c.logger.Debug("file uploaded", "filename", files[0].Filename, "size", fmt.Sprintf("%d bytes", files[0].Size), "content-type", meta.Mime)

	fileData := file.File{
		Name: 	  meta.Name,
		File: 	  meta.File,
		Public:   meta.Public,
		Mime:     meta.Mime,
		Grant:    meta.Grant,
		Size:     int(files[0].Size),
		Metadata: convertToStringMap(documentData),
		Reader:   r,
		User:     token.Login,
	}

	res, err := c.filesService.UploadFiles(fileData)
	if err != nil {
		c.responseBuilder.Error(ctx, err)
		return
	}

	uploadFileResponseData := uploadDataResponse{
			Json: res,
			File: res.Name,
	}

	c.responseBuilder.Ok(ctx, nil, uploadFileResponseData)
}

// @Summary Get documents list
// @Description Get filtered and paginated list of documents. Returns own documents if login not specified.
// @Tags docs
// @Produce json
// @Param login query string false "User login filter (optional - returns own documents if not specified)"
// @Param key query string false "Column name for filtering (optional)"
// @Param value query string false "Filter value (optional)"
// @Param limit query int false "Number of documents to return (optional)" minimum(1) maximum(1000) default(50)
// @Success 200 {object} getFilesResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/docs [get]
//
// @Summary Check documents availability
// @Description Check if documents exist with given filters (HEAD request). Returns same headers as GET but without body.
// @Tags docs
// @Success 200 "Documents exist"
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/docs [head]
func (c *Controller) GetFiles(ctx *gin.Context) {
	token := c.utils.GetTokenFromHeader(ctx)
	if token == nil {
		return
	}

	login := ctx.Query("login")
	if login == "" {
		login = token.Login
	}
	key := ctx.Query("key")
	value := ctx.Query("value")
	limitStr := ctx.Query("limit")
	var limit int
	var err error
	if limitStr == "" {
		limit = 0
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("invalid limit value"))
			return
		}
	}

	filter := contracts.NewFilterData(value, key, limit)
	files, err := c.filesService.GetFilesByUser(login, *filter)
	if err != nil {
		c.responseBuilder.Error(ctx, err)
		return
	}

	actualEtag := generateCollectionETag(files, *filter)
	if ctx.Request.Header.Get("If-None-Match") == actualEtag {
		ctx.Status(http.StatusNotModified)
		return
	}

    ctx.Header("ETag", "\""+actualEtag+"\"")
    ctx.Header("Cache-Control", "public, max-age=43200")
    ctx.Header("X-File-Count", strconv.Itoa(len(files)))

	filesResponse := filesDataResponse{
		Docs: files,
	}

	switch ctx.Request.Method {
	case "GET":
		c.responseBuilder.Ok(ctx, nil, filesResponse)
	case "HEAD":
		ctx.Status(http.StatusOK)
	}
}

// @Summary Get document by ID
// @Description Get single document by ID. Returns file content or JSON data depending on document type.
// @Tags docs
// @Produce json,octet-stream
// @Param id path string true "Document ID"
// @Success 200 {object} getFileResponse "JSON document retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/docs/{id} [get]
//
// @Summary Check document availability
// @Description Check if document exists and get headers (HEAD request). Returns same headers as GET but without body.
// @Tags docs
// @Param id path string true "Document ID"
// @Param token query string true "Authentication token"
// @Success 200 "Documents exist"
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/docs/{id} [head]
func (c *Controller) GetFile(ctx *gin.Context) {
	token := c.utils.GetTokenFromHeader(ctx)
	if token == nil {
		return
	}

	fileID := ctx.Param("docs_id")
	fileData, err := c.filesService.GetFileByID(fileID, token.Login)
	if err != nil {
		c.responseBuilder.Error(ctx, err)
		return
	}

	actualEtag := generateETag(*fileData)
	if ctx.Request.Header.Get("If-None-Match") == actualEtag {
		ctx.Status(http.StatusNotModified)
		return
	}

    ctx.Header("Content-Length", strconv.FormatInt(int64(fileData.Size), 10))
    ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileData.Name))
    ctx.Header("Cache-Control", "public, max-age=3600")
    ctx.Header("Last-Modified", fileData.CreatedAt.String())
    ctx.Header("ETag", "\""+actualEtag+"\"")

	switch ctx.Request.Method {
	case "GET":
		reader := fileData.Reader
		fileData.Reader = nil

		if !fileData.File {
			c.responseBuilder.Ok(ctx, nil, fileData)
			return
		}

		ctx.Header("Content-Type", fileData.Mime) 
		_, err = io.Copy(ctx.Writer, reader)
		if err != nil {
			c.logger.Error("failed to send file", "error", err)
			ctx.Status(http.StatusInternalServerError)
			return
		}

		ctx.Status(http.StatusOK)
		return
	case "HEAD":
		ctx.Status(http.StatusOK)
	}
}

// @Summary Delete document
// @Description Delete document by ID
// @Tags docs
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} deleteFileResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/docs/{id} [delete]
func (c *Controller) DeleteFile(ctx *gin.Context) {
	token := c.utils.GetTokenFromHeader(ctx)
	if token == nil {
		return
	}

	fileID := ctx.Param("docs_id")
	if fileID == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("docs_id filed is required"))
		return
	}

	fileData, err := c.filesService.DeleteFile(fileID, token.Login)
	if err != nil {
		c.responseBuilder.Error(ctx, err)
		return
	}

	isDeletedFile := map[string]bool{
		fileData.ID: true,
	}

	c.responseBuilder.Ok(ctx, isDeletedFile, nil)
}


func convertToStringMap(input map[string]any) map[string]string {
    result := make(map[string]string)
    for key, value := range input {
        switch v := value.(type) {
        case string:
            result[key] = v
        case fmt.Stringer:
            result[key] = v.String()
        default:
            result[key] = fmt.Sprintf("%v", v)
        }
    }
    return result
}

func generateETag(file file.File) string {
    data := fmt.Sprintf("%s-%s-%d-%t-%v",
        file.ID,
        file.Name,
        file.Size,
        file.Public,
        file.CreatedAt.Unix(),
    )
    
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

func generateCollectionETag(files []file.File, filter contracts.FilterData) string {
    components := []string{
        fmt.Sprintf("count:%d", len(files)),
        fmt.Sprintf("total:%d", len(files)),
        fmt.Sprintf("filter:%s:%s", filter.Key, filter.Value),
        fmt.Sprintf("limit:%d", filter.Limit),
    }

    fileHashes := make([]string, len(files))
    for i, file := range files {
        fileHashes[i] = generateETag(file)
    }
    
    sort.Strings(fileHashes)
    components = append(components, "files:"+strings.Join(fileHashes, ","))
    content := strings.Join(components, "|")
    hash := sha256.Sum256([]byte(content))
    return `"` + hex.EncodeToString(hash[:12]) + `"`
}