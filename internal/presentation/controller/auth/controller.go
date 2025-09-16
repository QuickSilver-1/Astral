package authcontroller

import (
	"astral/internal/domain/contracts"
	"astral/internal/presentation/controller/utils"
	"astral/internal/presentation/response"
	"log/slog"
)

type Controller struct {
	logger           *slog.Logger
	responseBuilder  *response.ResponseBuilder
	authService      contracts.AuthInterface
	adminToken 		 string
	utils            utils.Utils
}

func NewController(
	logger 			*slog.Logger,
	responseBuilder *response.ResponseBuilder,
	auth 			contracts.AuthInterface,
	token			string,
	utils           utils.Utils,
) *Controller {
	logger = logger.With("controller", "auth")
	return &Controller{
		logger:           logger,
		responseBuilder:  responseBuilder,
		authService: 	  auth,
		adminToken:		  token,
		utils:            utils,
	}
}
