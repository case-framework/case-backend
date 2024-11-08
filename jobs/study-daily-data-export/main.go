package main

import (
	"context"
	"log/slog"
	"os"
	"path"
	"time"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func main() {
	slog.Info("Starting study daily data export job")
	start := time.Now()

	for _, source := range conf.ResponseExports.Sources {
		runResponseExportsForSource(source.InstanceID, source.StudyKey, source.SurveyKeys)
	}

	if err := studyDBService.DBClient.Disconnect(context.Background()); err != nil {
		slog.Error("Error closing DB connection", slog.String("error", err.Error()))
	}
	slog.Info("Study daily data export job completed", slog.String("duration", time.Since(start).String()))
}

func runResponseExportsForSource(instanceID string, studyKey string, surveyKeys []string) {
	// ensure there is a folder path for the source (export_path/instance_id/study_key)
	relativeFolderPath := path.Join(instanceID, studyKey)
	exportFolderPathForSource := path.Join(conf.ResponseExports.ExportPath, relativeFolderPath)
	if _, err := os.Stat(exportFolderPathForSource); os.IsNotExist(err) {
		// create folder
		err = os.MkdirAll(exportFolderPathForSource, os.ModePerm)
		if err != nil {
			slog.Error("Error creating export path", slog.String("error", err.Error()))
			return
		}
		slog.Info("Created export path", slog.String("path", exportFolderPathForSource))
	}

	slog.Debug(exportFolderPathForSource)

	// TODO: remove old files (keep only the last retention_days, but at least yesterday and today)

	for _, surveyKey := range surveyKeys {
		// TODO: fetch survey infos
		_, err := getSurveyInfo(instanceID, studyKey, surveyKey)
		if err != nil {
			continue
		}
		// TODO: init response parser

		for i := 0; i < conf.ResponseExports.RetentionDays; i++ {
			// TODO: generate day string (yesterday - 1 - i)

			// TODO: check if file exists
			// TODO: if not, create file (with content)
		}

		// TODO: override / recreate yesterday file
		// TODO: override / recreate today file
	}
}

func getSurveyInfo(instanceID string, studyKey string, surveyKey string) (surveyInfos []*studyTypes.Survey, err error) {
	surveyInfos, err = studyDBService.GetSurveyVersions(instanceID, studyKey, surveyKey)
	if err != nil {
		slog.Error("Error getting survey versions", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey), slog.String("error", err.Error()))
		return
	}
	return
}
