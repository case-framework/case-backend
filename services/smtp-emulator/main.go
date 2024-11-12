package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

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
}
