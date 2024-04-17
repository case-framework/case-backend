package main

import (
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/services/smtp-bridge/apihandlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	sc "github.com/case-framework/case-backend/pkg/smtp-client"
)

var conf config

func main() {
	// Start webserver
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		// AllowAllOrigins: true,
		AllowOrigins:     conf.GinConfig.AllowOrigins,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Content-Length"},
		ExposeHeaders:    []string{"Authorization", "Content-Type", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	smtpClients, err := sc.NewSmtpClients(conf.SMTPServerConfig.LowPrio)
	if err != nil {
		slog.Error("Error creating SMTP clients", slog.String("error", err.Error()))
		panic("Error creating SMTP clients")
	}
	highPrioSmtpClients, err := sc.NewSmtpClients(conf.SMTPServerConfig.HighPrio)
	if err != nil {
		slog.Error("Error creating high priority SMTP clients", slog.String("error", err.Error()))
		panic("Error creating high priority SMTP clients")
	}

	// Add handlers
	router.GET("/", apihandlers.HealthCheckHandle)
	root := router.Group("/")
	apiModule := apihandlers.NewHTTPHandler(
		conf.ApiKeys,
		highPrioSmtpClients,
		smtpClients,
	)

	apiModule.AddRoutes(root)

	if conf.GinConfig.DebugMode {
		apihelpers.WriteRoutesToFile(router, "smtp-bridge-api-routes.txt")
	}

	slog.Info("Starting SMTP Bridge API on port " + conf.GinConfig.Port)
	err = router.Run(":" + conf.GinConfig.Port)
	if err != nil {
		slog.Error("Exited SMTP Bridge API", slog.String("error", err.Error()))
		return
	}
}
