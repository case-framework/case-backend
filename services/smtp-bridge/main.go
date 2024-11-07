package main

import (
	"log/slog"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/services/smtp-bridge/apihandlers"

	"github.com/gin-gonic/gin"

	sc "github.com/case-framework/case-backend/pkg/smtp-client"
)

var conf config

func main() {
	// Start webserver
	router := gin.Default()

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
