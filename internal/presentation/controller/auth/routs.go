package authcontroller

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

// @Summary Register auth management routes
// @Description Group of endpoints for working with authorization
func (r *Router) RegisterRoutes(auth *gin.RouterGroup, authSecure *gin.RouterGroup) {
	auth.POST("/register", r.controller.Register)
	auth.POST("/auth", r.controller.Login)
	authSecure.DELETE("/auth/:token_id", r.controller.DeleteSession)
}
