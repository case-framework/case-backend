package studyutils

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	studydb "github.com/case-framework/case-backend/pkg/db/study"
)

func CleanUpOrphanedTaskResults(
	instanceID string,
	studyDBService *studydb.StudyDBService,
	rootPath string,
) {
	slog.Info("Cleaning up orphaned task results (old files)", slog.String("instanceID", instanceID))

	folder := path.Join(rootPath, instanceID)

	// get all files in folder recursively
	files, err := filepath.Glob(folder + "/**/*")
	if err != nil {
		slog.Error("Failed to get files", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		return
	}

	for _, file := range files {
		// check if file is a task result file
		relativeFilepath := (strings.TrimPrefix(file, path.Clean(rootPath)))[1:]

		_, err := studyDBService.GetTaskByFilename(instanceID, relativeFilepath)
		if err != nil {
			slog.Info("Task for file not found, removing file", slog.String("reason", err.Error()), slog.String("instanceID", instanceID), slog.String("file", relativeFilepath))
			err = os.Remove(file)
			if err != nil {
				slog.Error("Failed to remove file", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("file", relativeFilepath))
			}
			continue
		}
	}
}
