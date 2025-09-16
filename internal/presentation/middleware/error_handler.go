package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standard API error response
// @Description Standard error response object
type ErrorResponse struct {
	Code    string      `json:"code" example:"PERMISSION_DENIED"`
	Message string      `json:"message" example:"You don't have permission to access this resource"`
	Details interface{} `json:"details,omitempty"`
}

func (m *Middleware) ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		m.logger.Info("request error",
				slog.String("path", ctx.Request.URL.Path),
				slog.String("method", ctx.Request.Method),
				slog.String("client_ip", ctx.ClientIP()),
			)
			
		ctx.Next()
		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last().Err

			m.logger.Error("request error",
				slog.String("error", err.Error()),
				slog.String("path", ctx.Request.URL.Path),
				slog.String("method", ctx.Request.Method),
				slog.String("client_ip", ctx.ClientIP()),
			)

			m.response.Error(ctx, err)
		}
	}
}
