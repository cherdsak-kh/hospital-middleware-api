package http

import (
	"net/http"

	"github.com/cherdsak-kh/hospital-middleware-api/config"
	_ "github.com/cherdsak-kh/hospital-middleware-api/docs"
	"github.com/cherdsak-kh/hospital-middleware-api/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter sets up Gin engine routes, middleware, and handlers
func SetupRouter(
	cfg *config.Config,
	staffHandler *StaffHandler,
	patientHandler *PatientHandler,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Simple CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "hospital-middleware-api",
			"status":  "running",
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "hospital-middleware-api",
		})
	})

	// Swagger API documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes (Staff registration and login)
	staffGroup := router.Group("/staff")
	{
		staffGroup.POST("/create", staffHandler.CreateStaff)
		staffGroup.POST("/login", staffHandler.LoginStaff)
	}

	// Protected routes (Requires valid JWT token with hospital context)
	patientGroup := router.Group("/patient")
	patientGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		patientGroup.GET("/search", patientHandler.SearchPatients)
		patientGroup.POST("/search", patientHandler.SearchPatients)
	}

	return router
}
