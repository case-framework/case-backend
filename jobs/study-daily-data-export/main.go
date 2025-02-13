package main

import (
	"context"
	"log/slog"
	"time"
)

func main() {
	slog.Info("Starting study daily data export job")
	start := time.Now()

	for _, rExpTask := range conf.ResponseExports.ExportTasks {
		slog.Info("Running response export task", slog.String("instanceID", rExpTask.InstanceID), slog.String("studyKey", rExpTask.StudyKey))
		runResponseExportsForTask(rExpTask)
	}

	cleanUpConfidentialResponsesExports()
	for _, rExpTask := range conf.ConfidentialResponsesExports.ExportTasks {
		slog.Info("Running confidential responses export task", slog.String("instanceID", rExpTask.InstanceID), slog.String("studyKey", rExpTask.StudyKey), slog.String("name", rExpTask.Name))
		runConfidentialResponsesExportsForTask(rExpTask)
	}

	if err := studyDBService.DBClient.Disconnect(context.Background()); err != nil {
		slog.Error("Error closing DB connection", slog.String("error", err.Error()))
	}
	slog.Info("Study daily data export job completed", slog.String("duration", time.Since(start).String()))
}
