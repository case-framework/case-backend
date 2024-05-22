package main

import (
	"log/slog"
	"time"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	slog.Info("Starting study timer job")
	start := time.Now()

	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start handling study timer for instance", slog.String("instanceID", instanceID))
		studies, err := studyDBService.GetStudies(instanceID, studyTypes.STUDY_STATUS_ACTIVE, false)
		if err != nil {
			slog.Error("Failed to get studies", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
			continue
		}

		for _, study := range studies {
			updateStudyStats(instanceID, study)
			performStudyTimer(instanceID, study)
		}
	}

	slog.Info("Study timer job completed", slog.String("duration", time.Since(start).String()))
}

func updateStudyStats(instanceID string, study studyTypes.Study) {
	activeCount, err := studyDBService.GetParticipantCount(instanceID, study.Key, bson.M{
		"studyStatus": studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE,
	})
	if err != nil {
		slog.Error("Failed to get active participant count", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}

	temporaryCount, err := studyDBService.GetParticipantCount(instanceID, study.Key, bson.M{
		"studyStatus": studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY,
	})
	if err != nil {
		slog.Error("Failed to get temporary participant count", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}

	responseCount, err := studyDBService.GetResponsesCount(instanceID, study.Key, bson.M{})
	if err != nil {
		slog.Error("Failed to get response count", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}

	stats := studyTypes.StudyStats{
		ParticipantCount:     activeCount,
		TempParticipantCount: temporaryCount,
		ResponseCount:        responseCount,
	}

	err = studyDBService.UpdateStudyStats(instanceID, study.Key, stats)
	if err != nil {
		slog.Error("Failed to update study stats", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

func performStudyTimer(instanceID string, study studyTypes.Study) {

}
