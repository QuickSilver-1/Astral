package http

import (
	_ "astral/docs"
	"astral/env"
	"astral/internal/domain/contracts"
	authcontroller "astral/internal/presentation/controller/auth"
	filescontroller "astral/internal/presentation/controller/files"
	"astral/internal/presentation/controller/utils"
	"astral/internal/presentation/middleware"
	"astral/internal/presentation/response"
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	logger 			  *slog.Logger
	fileService 	  contracts.FilesInterface
	authService 	  contracts.AuthInterface
	validationService contracts.ValidationInterface
	enviroments       env.Env
}

func NewHandler(
	logger 				*slog.Logger,
	fileService 		contracts.FilesInterface,
	authService 		contracts.AuthInterface,
	validationService 	contracts.ValidationInterface,
	enviroments         env.Env,
) *Handler {
	return &Handler{
		logger: 		   logger,
		fileService:	   fileService,
		authService: 	   authService,
		validationService: validationService,
		enviroments:       enviroments,
	}
}

// @title           			Astral File Service
// @version         			1.0
// @description     			Backend server for file service
// @host            			localhost:8080
// @BasePath        			/api
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
func (c *Handler) InitRouts(environments *env.Env) *gin.Engine {
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-type"},
		ExposeHeaders:    []string{"Content-Length, Content-Type, ETag, Last-Modified"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, ETag, Last-Modified")
			c.Header("Access-Control-Max-Age", "43200")
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	rBuilder := response.NewResponseBuilder()

	router.Use(gin.Recovery())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	middlewareController := middleware.NewAuthMiddleware(rBuilder, c.authService, c.logger)
	api := router.Group("/api")
	router.Use(middlewareController.UserIdentity())
	router.Use(middlewareController.ErrorHandler())
	secureApi := router.Group("/api")

	utilsController := utils.NewUtils(c.logger, c.authService, *rBuilder)

	authController := authcontroller.NewController(c.logger, rBuilder, c.authService, c.enviroments.AdminToken, *utilsController)
	authRouter := authcontroller.NewRouter(authController)
	authRouter.RegisterRoutes(api, secureApi)

	filesController := filescontroller.NewController(c.logger, rBuilder, c.fileService, c.enviroments.AdminToken)
	filesRouter := filescontroller.NewRouter(filesController)
	filesRouter.RegisterRoutes(secureApi)

	return router
}
