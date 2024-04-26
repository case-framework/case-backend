package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/services/participant-api/apihandlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var conf ParticipantApiConfig

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

	// Add handlers
	router.GET("/", apihandlers.HealthCheckHandle)
	v1Root := router.Group("/v1")

	v1APIHandlers := apihandlers.NewHTTPHandler(
		conf.UserManagementConfig.ParticipantUserJWTConfig.SignKey,
		studyDBService,
		participantUserDBService,
		globalInfosDBService,
		messagingDBService,
		conf.AllowedInstanceIDs,
		conf.StudyConfigs.GlobalSecret,
		conf.FilestorePath,
		conf.UserManagementConfig.MaxNewUsersPer5Minutes,
		apihandlers.TTLs{
			AccessToken:                   conf.UserManagementConfig.ParticipantUserJWTConfig.ExpiresIn,
			EmailContactVerificationToken: conf.UserManagementConfig.EmailContactVerificationTokenTTL,
		},
	)
	v1APIHandlers.AddParticipantAuthAPI(v1Root)

	if conf.GinConfig.DebugMode {
		apihelpers.WriteRoutesToFile(router, "participant-api-routes.txt")
	}

	// Start the server
	slog.Info("Starting Participant API on port " + conf.GinConfig.Port)
	if !conf.GinConfig.MTLS.Use {
		err := router.Run(":" + conf.GinConfig.Port)
		if err != nil {
			slog.Error("Exited Participant API", slog.String("error", err.Error()))
			return
		}
	} else {
		// Create tls config for mutual TLS
		tlsConfig, err := apihelpers.LoadTLSConfig(conf.GinConfig.MTLS.CertificatePaths)
		if err != nil {
			slog.Error("Error loading TLS config.", slog.String("error", err.Error()))
			return
		}

		server := &http.Server{
			Addr:      ":" + conf.GinConfig.Port,
			Handler:   router,
			TLSConfig: tlsConfig,
		}

		err = server.ListenAndServeTLS(conf.GinConfig.MTLS.CertificatePaths.ServerCertPath, conf.GinConfig.MTLS.CertificatePaths.ServerKeyPath)
		if err != nil {
			slog.Error("Exited Participant API", slog.String("error", err.Error()))
			return
		}
	}

}
