/*package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)
var conf config

func main() {
	// Start webserver
	router := gin.Default()

	// Add handlers
	router.POST("/send-email", sendEmail)
	err := router.Run(":8090")

	slog.Info("Starting SMTP Bridge Emulator API on port 8090")
	if err != nil {
		slog.Error("Exited SMTP Bridge Emulator API", slog.String("error", err.Error()))
		return
	}
}*/

package main

import (
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/services/smtp-bridge-emulator/apihandlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// Add handlers
	router.GET("/", apihandlers.HealthCheckHandle)
	root := router.Group("/")
	apiModule := apihandlers.NewHTTPHandler(
		conf.ApiKeys)

	apiModule.AddRoutes(root)

	if conf.GinConfig.DebugMode {
		apihelpers.WriteRoutesToFile(router, "smtp-bridge-emulator-api-routes.txt")
	}

	slog.Info("Starting SMTP Bridge emulator API on port " + conf.GinConfig.Port)
	err := router.Run(":" + conf.GinConfig.Port)
	if err != nil {
		slog.Error("Exited SMTP Bridge emulator API", slog.String("error", err.Error()))
		return
	}
}
