package main

import (
	"context"
	"log/slog"
	"time"
)

func main() {
	slog.Info("Starting study daily data export job")
	start := time.Now()

	// TODO: implement daily data export

	if err := studyDBService.DBClient.Disconnect(context.Background()); err != nil {
		slog.Error("Error closing DB connection", slog.String("error", err.Error()))
	}
	slog.Info("Study daily data export job completed", slog.String("duration", time.Since(start).String()))
}
