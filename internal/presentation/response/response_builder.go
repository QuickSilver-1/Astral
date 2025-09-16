package response

import (
	controllererrors "astral/internal/presentation/controller/errors"
	authrepo "astral/internal/repository/auth"
	filesrepo "astral/internal/repository/files"
	authservice "astral/internal/services/authorization"
	fileservice "astral/internal/services/files"
	validationservice "astral/internal/services/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseBuilder struct{}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{}
}

type ErrorResponse struct {
	Code int 	`json:"code,omitempty"`
	Text string `json:"text,omitempty"`
}

type Response struct {
	Error 	  ErrorResponse 	`json:"error,omitzero"`
	Resp      any               `json:"response,omitempty"`
	Data      any 				`json:"data,omitempty" swaggertype:"object"`
}

func (f *ResponseBuilder) Ok(ctx *gin.Context, response, data any) {
	ctx.JSON(http.StatusOK, Response{
		Resp: response,
		Data: data,
	})
}

func (f *ResponseBuilder) Error(ctx *gin.Context, err error) {
	switch err {
	case authrepo.ErrUserAlreadyExists:
		ctx.AbortWithStatusJSON(http.StatusConflict, getErrorResponse(http.StatusConflict, err.Error()))
	case authrepo.ErrNoRows:
		ctx.AbortWithStatusJSON(http.StatusNotFound, getErrorResponse(http.StatusNotFound, err.Error()))
	case filesrepo.ErrFileNotFound:
		ctx.AbortWithStatusJSON(http.StatusNotFound, getErrorResponse(http.StatusNotFound, err.Error()))
	case fileservice.ErrAccessDenied:
		ctx.AbortWithStatusJSON(http.StatusForbidden, getErrorResponse(http.StatusForbidden, err.Error()))
	case authservice.ErrAccessDenied:
		ctx.AbortWithStatusJSON(http.StatusForbidden, getErrorResponse(http.StatusForbidden, err.Error()))
	case authservice.ErrInvalidToken:
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, getErrorResponse(http.StatusUnauthorized, err.Error()))
		
	default:
		switch err.(type) {
		case controllererrors.ErrInvalidInputData:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, getErrorResponse(http.StatusBadRequest, err.Error()))
		case filesrepo.ErrFileUpload:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, getErrorResponse(http.StatusBadRequest, err.Error()))
		case validationservice.ErrValidationUserData:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, getErrorResponse(http.StatusBadRequest, err.Error()))

		default:
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, getErrorResponse(http.StatusInternalServerError, "Internal server error"))
		}
	}
}

func getErrorResponse(code int, text string) ErrorResponse {
	return ErrorResponse{
		Code: code,
		Text: text,
	}
}