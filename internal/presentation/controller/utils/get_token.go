package utils

import (
	"astral/internal/domain/user"
	authservice "astral/internal/services/authorization"
	"fmt"

	"github.com/gin-gonic/gin"
)

func (c *Utils) GetTokenFromHeader(ctx *gin.Context) *user.Token {
	tokenAny, exists := ctx.Get("user")
	if !exists {
		c.logger.Warn("userID not found in context")
		c.responseBuilder.Error(ctx, authservice.ErrInvalidToken)
		return nil
	}

	token, ok := tokenAny.(*user.Token)
	if !ok {
		c.logger.Warn("invalid identity type in context", "type", fmt.Sprintf("%T", tokenAny))
		c.responseBuilder.Error(ctx, authservice.ErrInvalidToken)
		return nil
	}

	return token
}