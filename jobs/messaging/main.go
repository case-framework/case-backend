package main

import (
	"log/slog"
	"time"
)

func main() {
	slog.Info("Starting messaging job")
	start := time.Now()

	// TODO: Implement messaging job

	slog.Info("Messaging job completed", slog.String("duration", time.Since(start).String()))
}
