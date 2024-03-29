package main

import (
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/services/smtp-bridge/apihandlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var conf config

func main() {
	// Start webserver
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		// AllowAllOrigins: true,
		AllowOrigins:     conf.AllowOrigins,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Content-Length"},
		ExposeHeaders:    []string{"Authorization", "Content-Type", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add handlers
	router.GET("/", apihandlers.HealthCheckHandle)
	root := router.Group("/")
	apiModule := apihandlers.NewHTTPHandler(
		[]string{},
	)

	apiModule.AddRoutes(root)

	if conf.GinDebugMode {
		apihelpers.WriteRoutesToFile(router, "smtp-bridge-api-routes.txt")
	}

	slog.Info("Starting SMTP Bridge API on port " + conf.Port)
	err := router.Run(":" + conf.Port)
	if err != nil {
		slog.Error("Exited SMTP Bridge API", slog.String("error", err.Error()))
		return
	}
}
