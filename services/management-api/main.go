package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
	"github.com/case-framework/case-backend/pkg/db/messaging"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/case-framework/case-backend/services/management-api/apihandlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var conf Config

func main() {
	// Connect to DBs
	muDBService, err := muDB.NewManagementUserDBService(conf.ManagementUserDBConfig)
	if err != nil {
		slog.Error("Error connecting to Management User DB", slog.String("error", err.Error()))
		return
	}
	messagingDBService, err := messaging.NewMessagingDBService(conf.MessagingDBConfig)
	if err != nil {
		slog.Error("Error connecting to Messaging DB", slog.String("error", err.Error()))
		return
	}

	studyDBService, err := studyDB.NewStudyDBService(conf.StudyDBConfig)
	if err != nil {
		slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
		return
	}

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
	v1Root := router.Group("/v1")

	v1APIHandlers := apihandlers.NewHTTPHandler(
		conf.ManagementUserJWTSignKey,
		conf.ManagementUserJWTExpiresIn,
		muDBService,
		messagingDBService,
		studyDBService,
		conf.AllowedInstanceIDs,
		conf.StudyGlobalSecret,
		conf.FilestorePath,
	)
	v1APIHandlers.AddManagementAuthAPI(v1Root)
	v1APIHandlers.AddUserManagementAPI(v1Root)
	v1APIHandlers.AddMessagingServiceAPI(v1Root)
	v1APIHandlers.AddStudyManagementAPI(v1Root)

	if conf.GinDebugMode {
		apihelpers.WriteRoutesToFile(router, "management-api-routes.txt")
	}

	// Start the server
	slog.Info("Starting Management API on port " + conf.Port)
	if !conf.UseMTLS {
		err := router.Run(":" + conf.Port)
		if err != nil {
			slog.Error("Exited Management API", slog.String("error", err.Error()))
			return
		}
	} else {
		// Create tls config for mutual TLS
		tlsConfig, err := apihelpers.LoadTLSConfig(conf.CertificatePaths)
		if err != nil {
			slog.Error("Error loading TLS config.", slog.String("error", err.Error()))
			return
		}

		server := &http.Server{
			Addr:      ":" + conf.Port,
			Handler:   router,
			TLSConfig: tlsConfig,
		}

		err = server.ListenAndServeTLS(conf.CertificatePaths.ServerCertPath, conf.CertificatePaths.ServerKeyPath)
		if err != nil {
			slog.Error("Exited Management API", slog.String("error", err.Error()))
			return
		}
	}
}
