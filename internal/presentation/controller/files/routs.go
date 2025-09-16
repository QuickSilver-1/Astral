package filescontroller

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	controller *Controller
}

func NewRouter(controller *Controller) *Router {
	return &Router{
		controller: controller,
	}
}

// @Summary Register files management routes
// @Description Group of endpoints for working with files
func (r *Router) RegisterRoutes(files *gin.RouterGroup) {
	files.POST("/docs", r.controller.UploadFile)

	files.GET("/docs", r.controller.GetFiles)
	files.HEAD("/docs", r.controller.GetFiles)

	files.GET("/docs/:docs_id", r.controller.GetFile)
	files.HEAD("/docs/:docs_id", r.controller.GetFile)
	files.DELETE("/docs/:docs_id", r.controller.DeleteFile)
}
