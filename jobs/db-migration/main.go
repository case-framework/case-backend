package main

import (
	"log/slog"
	"time"
)

func main() {
	dropIndexes()

	createIndexes()

	migrationTasks()
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
