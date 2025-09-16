package utils

import (
	"astral/internal/domain/contracts"
	"astral/internal/presentation/response"
	"log/slog"
)

type Utils struct {
	responseBuilder response.ResponseBuilder
	authservice   	contracts.AuthInterface
	logger 		    *slog.Logger
}

func NewUtils(logger *slog.Logger, authService contracts.AuthInterface, rBuilder response.ResponseBuilder) *Utils {
	return &Utils{
		responseBuilder: rBuilder,
		authservice: 	 authService,
		logger: 	     logger,
	}
}