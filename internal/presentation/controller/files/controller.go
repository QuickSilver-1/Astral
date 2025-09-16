package filescontroller

import (
	"astral/internal/domain/contracts"
	"astral/internal/presentation/controller/utils"
	"astral/internal/presentation/response"
	"log/slog"
)

type Controller struct {
	logger           *slog.Logger
	responseBuilder  *response.ResponseBuilder
	filesService     contracts.FilesInterface
	adminToken 		 string
	utils            utils.Utils
}

func NewController(
	logger 			*slog.Logger,
	responseBuilder *response.ResponseBuilder,
	files 			contracts.FilesInterface,
	token			string,
) *Controller {
	logger = logger.With("controller", "files")
	return &Controller{
		logger:           logger,
		responseBuilder:  responseBuilder,
		adminToken:		  token,
		filesService:     files,
	}
}
