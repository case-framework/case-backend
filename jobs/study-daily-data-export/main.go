package main

import (
	"context"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	surveydefinition "github.com/case-framework/case-backend/pkg/study/exporter/survey-definition"
	surveyresponses "github.com/case-framework/case-backend/pkg/study/exporter/survey-responses"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	slog.Info("Starting study daily data export job")
	start := time.Now()

	for _, rExpTask := range conf.ResponseExports.ExportTasks {
		runResponseExportsForTask(rExpTask)
	}

	if err := studyDBService.DBClient.Disconnect(context.Background()); err != nil {
		slog.Error("Error closing DB connection", slog.String("error", err.Error()))
	}
	slog.Info("Study daily data export job completed", slog.String("duration", time.Since(start).String()))
}

func runResponseExportsForTask(rExpTask ResponseExportTask) {
	// ensure there is a folder path for the source (export_path/instance_id/study_key)
	relativeFolderPath := path.Join(rExpTask.InstanceID, rExpTask.StudyKey)
	exportFolderPathForSource := path.Join(conf.ExportPath, relativeFolderPath)
	if _, err := os.Stat(exportFolderPathForSource); os.IsNotExist(err) {
		// create folder
		err = os.MkdirAll(exportFolderPathForSource, os.ModePerm)
		if err != nil {
			slog.Error("Error creating export path", slog.String("error", err.Error()))
			return
		}
		slog.Info("Created export path", slog.String("path", exportFolderPathForSource))
	}

	// remove old files (keep only the last retention_days, but at least yesterday and today)
	if err := cleanUpForSource(exportFolderPathForSource); err != nil {
		slog.Error("Error cleaning up old files", slog.String("error", err.Error()))
	}

	for _, surveyKey := range rExpTask.SurveyKeys {
		parser, err := initResponseParser(rExpTask.InstanceID, rExpTask.StudyKey, surveyKey, rExpTask.ShortKeys, rExpTask.Separator)
		if err != nil {
			continue
		}

		if conf.ResponseExports.OverrideOld {
			for i := 0; i < conf.ResponseExports.RetentionDays-1; i++ {
				targetDate := time.Now().Add(
					time.Duration(-(conf.ResponseExports.RetentionDays - i)) * time.Hour * 24,
				)
				generateExportForSurveyForTargetDate(rExpTask.InstanceID, rExpTask.StudyKey, surveyKey, rExpTask.ExportFormat, targetDate, exportFolderPathForSource, parser)
			}
		}

		// yesterday
		targetDate := time.Now().Add(
			time.Duration(-1 * time.Hour * 24),
		)
		generateExportForSurveyForTargetDate(rExpTask.InstanceID, rExpTask.StudyKey, surveyKey, rExpTask.ExportFormat, targetDate, exportFolderPathForSource, parser)

		// today
		targetDate = time.Now()
		generateExportForSurveyForTargetDate(rExpTask.InstanceID, rExpTask.StudyKey, surveyKey, rExpTask.ExportFormat, targetDate, exportFolderPathForSource, parser)
	}
}

func initResponseParser(instanceID string, studyKey string, surveyKey string, shortKeys bool, separator string) (parser *surveyresponses.ResponseParser, err error) {
	surveyVersions, err := surveydefinition.PrepareSurveyInfosFromDB(
		studyDBService,
		instanceID,
		studyKey,
		surveyKey,
		&surveydefinition.ExtractOptions{
			UseLabelLang: "",
			IncludeItems: nil,
			ExcludeItems: nil,
		},
	)
	if err != nil {
		slog.Error("failed to get survey versions", slog.String("error", err.Error()))
		return
	}
	extraCols := conf.ResponseExports.ExportTasks[0].ExtraCtxCols
	parser, err = surveyresponses.NewResponseParser(
		surveyKey,
		surveyVersions,
		shortKeys,
		nil,
		separator,
		&extraCols,
	)
	if err != nil {
		slog.Error("failed to create response parser", slog.String("error", err.Error()))
		return
	}
	return
}

func generateExportForSurveyForTargetDate(instanceID string, studyKey string, surveyKey string, format string, targetDate time.Time, exportPath string, parser *surveyresponses.ResponseParser) {
	fileName := responseFileName(targetDate, surveyKey, format)
	responseFilePath := filepath.Join(exportPath, fileName)

	if fileExists(responseFilePath) {
		slog.Debug("File already exists, overriding", slog.String("path", responseFilePath))
	}

	filter := bson.M{
		"key": surveyKey,
		"$and": bson.A{
			bson.M{"arrivedAt": bson.M{"$lte": endOfDay(targetDate).Unix()}},
			bson.M{"arrivedAt": bson.M{"$gte": startOfDay(targetDate).Unix()}},
		},
	}
	// count responses for target date and survey key --> if 0, skip
	count, err := studyDBService.GetResponsesCount(instanceID, studyKey, filter)
	if err != nil {
		slog.Error("Error getting responses count", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey), slog.String("error", err.Error()))
		return
	}
	if count == 0 {
		slog.Debug("No responses for target date and survey key, skipping", slog.String("targetDate", targetDate.Format("2006-01-02")), slog.String("surveyKey", surveyKey))
		return
	}

	file, err := os.Create(responseFilePath)
	if err != nil {
		slog.Error("failed to create export file", slog.String("error", err.Error()))
		return
	}

	defer file.Close()

	exporter, err := surveyresponses.NewResponseExporter(
		parser,
		file,
		format,
	)
	if err != nil {
		slog.Error("failed to create response exporter", slog.String("error", err.Error()))
		return
	}

	err = studyDBService.FindAndExecuteOnResponses(
		context.Background(),
		instanceID,
		studyKey,
		filter,
		bson.M{"arrivedAt": 1},
		false,
		func(dbService *studyDB.StudyDBService, r studyTypes.SurveyResponse, instanceID, studyKey string, args ...interface{}) error {
			err := exporter.WriteResponse(&r)
			if err != nil {
				return err
			}
			return nil
		},
		nil,
	)
	if err != nil {
		slog.Error("Error generating response export", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey), slog.String("error", err.Error()))
		return
	}

	err = exporter.Finish()
	if err != nil {
		slog.Error("failed to finish export", slog.String("error", err.Error()))
		return
	}
	slog.Info("Generated response export", slog.String("path", responseFilePath))
}

func cleanUpForSource(sourceDir string) error {
	cutoffDate := time.Now().Add(
		time.Duration(-(conf.ResponseExports.RetentionDays + 1)) * time.Hour * 24,
	)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Parse date from filename (assuming format YYYY-MM-DD##responses##..##..)
		basename := filepath.Base(path)
		parts := strings.Split(basename, "##")
		if len(parts) < 1 {
			return nil
		}
		datePart := parts[0]
		if len(datePart) < 10 {
			return nil
		}

		fileDate, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			return nil
		}

		if fileDate.Before(cutoffDate) {
			if err := os.Remove(path); err != nil {
				slog.Error("Failed to remove old file", slog.String("path", path), slog.String("error", err.Error()))
			}
		}

		return nil
	})
}
