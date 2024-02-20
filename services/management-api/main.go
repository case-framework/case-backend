package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/services/management-api/apihandlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	AllowOrigins     []string                    `json:"allow_origins"`
	Port             string                      `json:"port"`
	UseMTLS          bool                        `json:"use_mtls"`
	CertificatePaths apihelpers.CertificatePaths `json:"certificate_paths"`
}

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	conf := Config{
		AllowOrigins: []string{"*"},
		Port:         "8080",
		UseMTLS:      false,
	}

	apihandlers.HandlerTest()
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

	v1APIHandlers := apihandlers.NewHTTPHandler()
	v1APIHandlers.AddManagementAuthAPI(v1Root)

	// Start the server
	if !conf.UseMTLS {
		slog.Info("Starting Management API on port " + conf.Port)
		err := router.Run(":" + conf.Port)
		if err != nil {
			slog.Error("Exited Management API", err)
			return
		}
	} else {
		// Create tls config for mutual TLS
		tlsConfig, err := apihelpers.LoadTLSConfig(conf.CertificatePaths)
		if err != nil {
			slog.Error("Error loading TLS config.", err)
			return
		}

		server := &http.Server{
			Addr:      ":" + conf.Port,
			Handler:   router,
			TLSConfig: tlsConfig,
		}

		err = server.ListenAndServeTLS(conf.CertificatePaths.ServerCertPath, conf.CertificatePaths.ServerKeyPath)
		if err != nil {
			slog.Error("Exited Management API", err)
			return
		}
	}
}
