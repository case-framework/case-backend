package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/services/participant-api/apihandlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
)

var conf ParticipantApiConfig

func main() {
	studyDBService, err := studyDB.NewStudyDBService(conf.StudyDBConfig)
	if err != nil {
		slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
		return
	}

	userDbService, err := userDB.NewParticipantUserDBService(conf.ParticipantUserDBConfig)
	if err != nil {
		slog.Error("Error connecting to Participant User DB", slog.String("error", err.Error()))
		return
	}

	globalInfosDBService, err := globalinfosDB.NewGlobalInfosDBService(conf.GlobalInfosDBConfig)
	if err != nil {
		slog.Error("Error connecting to Global Infos DB", slog.String("error", err.Error()))
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
		conf.ParticipantUserJWTSignKey,
		conf.ParticipantUserJWTExpiresIn,
		studyDBService,
		userDbService,
		globalInfosDBService,
		conf.AllowedInstanceIDs,
		conf.StudyGlobalSecret,
		conf.FilestorePath,
		conf.MaxNewUsersPer5Minutes,
	)
	v1APIHandlers.AddParticipantAuthAPI(v1Root)

	if conf.GinDebugMode {
		apihelpers.WriteRoutesToFile(router, "participant-api-routes.txt")
	}

	// Start the server
	slog.Info("Starting Participant API on port " + conf.Port)
	if !conf.UseMTLS {
		err := router.Run(":" + conf.Port)
		if err != nil {
			slog.Error("Exited Participant API", slog.String("error", err.Error()))
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
			slog.Error("Exited Participant API", slog.String("error", err.Error()))
			return
		}
	}

}
