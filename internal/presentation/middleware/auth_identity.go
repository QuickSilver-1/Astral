package middleware

import (
	"astral/internal/domain/contracts"
	"astral/internal/presentation/response"
	authservice "astral/internal/services/authorization"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	logger        *slog.Logger
	response	  *response.ResponseBuilder
	authService   contracts.AuthInterface
}

func NewAuthMiddleware(response *response.ResponseBuilder, authService contracts.AuthInterface, logger *slog.Logger) *Middleware {
	return &Middleware{
		logger:      logger.With("type", "middleware.AuthMiddleware"),
		response: 	 response,
		authService: authService,
	}
}

func (m *Middleware) UserIdentity() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		tokenWords := strings.Split(token, " ")
		if len(tokenWords) < 2 {
			m.response.Error(ctx, authservice.ErrAccessDenied)
			return
		}

		token = tokenWords[1]
		tokenData, err := m.authService.Authorization(token)
		if err != nil {
			m.response.Error(ctx, err)
			return
		}

		ctx.Set("user", tokenData)
		ctx.Next()
	}
}