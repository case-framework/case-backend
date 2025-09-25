package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	dropIndexes()

	createIndexes()

	migrationTasks()

	getIndexes()
}

func dropIndexes() {
	switch conf.TaskConfigs.DropIndexes.StudyDB {
	case DropIndexesModeAll:
		studyDBService.DropAllIndexes()
	case DropIndexesModeDefaults:
		studyDBService.DropDefaultIndexes()
	}

	switch conf.TaskConfigs.DropIndexes.ParticipantUserDB {
	case DropIndexesModeAll:
		participantUserDBService.DropIndexes(true)
	case DropIndexesModeDefaults:
		participantUserDBService.DropIndexes(false)
	}

	switch conf.TaskConfigs.DropIndexes.ManagementUserDB {
	case DropIndexesModeAll:
		managementUserDBService.DropIndexes(true)
	case DropIndexesModeDefaults:
		managementUserDBService.DropIndexes(false)
	}

	switch conf.TaskConfigs.DropIndexes.GlobalInfosDB {
	case DropIndexesModeAll:
		globalInfosDBService.DropIndexes(true)
	case DropIndexesModeDefaults:
		globalInfosDBService.DropIndexes(false)
	}

	switch conf.TaskConfigs.DropIndexes.MessagingDB {
	case DropIndexesModeAll:
		messagingDBService.DropIndexes(true)
	case DropIndexesModeDefaults:
		messagingDBService.DropIndexes(false)
	}
}

func createIndexes() {
	if conf.TaskConfigs.CreateIndexes.StudyDB {
		studyDBService.CreateDefaultIndexes()
	}

	if conf.TaskConfigs.CreateIndexes.ParticipantUserDB {
		participantUserDBService.CreateDefaultIndexes()
	}

	if conf.TaskConfigs.CreateIndexes.ManagementUserDB {
		managementUserDBService.CreateDefaultIndexes()
	}

	if conf.TaskConfigs.CreateIndexes.GlobalInfosDB {
		globalInfosDBService.CreateDefaultIndexes()
	}

	if conf.TaskConfigs.CreateIndexes.MessagingDB {
		messagingDBService.CreateDefaultIndexes()
	}
}

func migrationTasks() {
	// Fix participant user contact infos
	if conf.TaskConfigs.MigrationTasks.ParticipantUserContactInfosFix {
		for _, instanceID := range participantUserDBService.InstanceIDs {
			start := time.Now()
			slog.Info("Fixing participant user contact infos", slog.String("instanceID", instanceID))
			err := participantUserDBService.FixFieldNameForContactInfos(instanceID)
			if err != nil {
				slog.Error("Error fixing participant user contact infos", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			}
			slog.Info("Participant user contact infos fixed", slog.String("instanceID", instanceID), slog.String("duration", time.Since(start).String()))
		}
	}
}

func getIndexes() {
	shouldGetIndexes := shouldGetIndexesForDBs()
	if shouldGetIndexes.StudyDB {
		indexes, err := studyDBService.GetIndexes()
		if err != nil {
			slog.Error("Error getting indexes for study DB", slog.String("error", err.Error()))
		} else {
			saveIndexAsJSON(indexes, conf.TaskConfigs.GetIndexes.StudyDB)
		}
	}

	if shouldGetIndexes.ParticipantUserDB {
		indexes, err := participantUserDBService.GetIndexes()
		if err != nil {
			slog.Error("Error getting indexes for participant user DB", slog.String("error", err.Error()))
		} else {
			saveIndexAsJSON(indexes, conf.TaskConfigs.GetIndexes.ParticipantUserDB)
		}
	}

	if shouldGetIndexes.ManagementUserDB {
		indexes, err := managementUserDBService.GetIndexes()
		if err != nil {
			slog.Error("Error getting indexes for management user DB", slog.String("error", err.Error()))
		} else {
			saveIndexAsJSON(indexes, conf.TaskConfigs.GetIndexes.ManagementUserDB)
		}
	}

	if shouldGetIndexes.GlobalInfosDB {
		indexes, err := globalInfosDBService.GetIndexes()
		if err != nil {
			slog.Error("Error getting indexes for global infos DB", slog.String("error", err.Error()))
		} else {
			saveIndexAsJSON(indexes, conf.TaskConfigs.GetIndexes.GlobalInfosDB)
		}
	}

	if shouldGetIndexes.MessagingDB {
		indexes, err := messagingDBService.GetIndexes()
		if err != nil {
			slog.Error("Error getting indexes for messaging DB", slog.String("error", err.Error()))
		} else {
			saveIndexAsJSON(indexes, conf.TaskConfigs.GetIndexes.MessagingDB)
		}
	}
}

func saveIndexAsJSON(indexes any, filename string) {
	if filename == "" || filename == "false" {
		slog.Warn("skipping index export: no file path configured")
		return
	}

	normalized, err := normalizeIndexes(indexes)
	if err != nil {
		slog.Error("Error normalizing indexes for JSON export", slog.String("error", err.Error()))
		normalized = indexes
	}

	jsonBytes, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		slog.Error("Error marshalling indexes to JSON", slog.String("error", err.Error()))
		return
	}
	jsonBytes = append(jsonBytes, '\n')

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("Error creating directory for index export", slog.String("path", dir), slog.String("error", err.Error()))
		return
	}

	if err := os.WriteFile(filename, jsonBytes, 0o644); err != nil {
		slog.Error("Error writing indexes JSON to file", slog.String("path", filename), slog.String("error", err.Error()))
		return
	}

	slog.Info("Indexes exported to JSON", slog.String("path", filename))
}

func normalizeIndexes(indexes any) (any, error) {
	switch v := indexes.(type) {
	case map[string]map[string][]bson.M:
		normalized := make(map[string]map[string][]map[string]any, len(v))
		for instanceID, collections := range v {
			collectionMap := make(map[string][]map[string]any, len(collections))
			for collectionName, idxList := range collections {
				converted := make([]map[string]any, len(idxList))
				for i, idx := range idxList {
					converted[i] = map[string]any(idx)
				}
				collectionMap[collectionName] = converted
			}
			normalized[instanceID] = collectionMap
		}
		return normalized, nil
	case map[string][]bson.M:
		normalized := make(map[string][]map[string]any, len(v))
		for collectionName, idxList := range v {
			converted := make([]map[string]any, len(idxList))
			for i, idx := range idxList {
				converted[i] = map[string]any(idx)
			}
			normalized[collectionName] = converted
		}
		return normalized, nil
	case map[string]any:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported index type %T", indexes)
	}
}
