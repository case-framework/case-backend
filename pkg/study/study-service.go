package study

import (
	"log/slog"

	studydb "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	"github.com/case-framework/case-backend/pkg/study/types"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	studyUtils "github.com/case-framework/case-backend/pkg/study/utils"
)

var (
	studyDBService *studydb.StudyDBService
	globalSecret   string
)

func Init(
	studyDB *studydb.StudyDBService,
	gSecret string,
	externalServices []studyengine.ExternalService,
) {
	studyDBService = studyDB
	globalSecret = gSecret
	studyengine.InitStudyEngine(studyDB, externalServices)
}

func OnProfileDeleted(instanceID, profileID string) {
	studies, err := studyDBService.GetStudies(instanceID, "", true)
	if err != nil {
		slog.Error("Error getting studies by status", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return
	}

	for _, study := range studies {
		studyKey := study.Key
		slog.Info("Processing study", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))

		participantID, confidentialID, err := computeParticipantIDs(study, profileID)
		if err != nil {
			slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
			continue
		}

		// compute participantIDs for study
		pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
		if err != nil {
			continue
		}

		currentEvent := studyengine.StudyEvent{
			Type:                                  studyengine.STUDY_EVENT_TYPE_LEAVE,
			InstanceID:                            instanceID,
			StudyKey:                              study.Key,
			ParticipantIDForConfidentialResponses: confidentialID,
		}

		// run study rules
		actionResult, err := getAndPerformStudyRules(instanceID, studyKey, pState, currentEvent)
		if err != nil {
			slog.Error("Error getting and performing study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
			continue
		}

		// save participant state
		actionResult.PState.StudyStatus = studyTypes.PARTICIPANT_STUDY_STATUS_ACCOUNT_DELETED

		_, err = studyDBService.SaveParticipantState(instanceID, study.Key, actionResult.PState)
		if err != nil {
			slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
			return
		}

		// save reports
		saveReports(
			instanceID,
			study.Key,
			actionResult.ReportsToCreate,
			studyengine.STUDY_EVENT_TYPE_LEAVE,
		)

		// delete confidential data
		_, err = studyDBService.DeleteConfidentialResponses(instanceID, studyKey, confidentialID, "")
		if err != nil {
			slog.Error("Error deleting confidential responses", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
			continue
		}

		// delete confidential ID to profileID mapping
		err = studyDBService.RemoveConfidentialIDMapEntriesForProfile(instanceID, profileID, studyKey)
		if err != nil {
			slog.Error("Error removing confidentialID map entries for profile", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("profileID", profileID), slog.String("error", err.Error()))
			continue
		}
	}
}

func getAndPerformStudyRules(instanceID, studyKey string, pState studyTypes.Participant, currentEvent studyengine.StudyEvent) (newState studyengine.ActionData, err error) {
	newState = studyengine.ActionData{
		PState:          pState,
		ReportsToCreate: map[string]types.Report{},
	}

	rulesObj, err := studyDBService.GetCurrentStudyRules(instanceID, studyKey)
	if err != nil {
		return
	}
	for _, rule := range rulesObj.Rules {
		newState, err = studyengine.ActionEval(rule, newState, currentEvent)
		if err != nil {
			return
		}
	}

	return newState, nil
}

func computeParticipantIDs(study studyTypes.Study, profileID string) (string, string, error) {
	pID, err := studyUtils.ProfileIDtoParticipantID(profileID, globalSecret, study.SecretKey, study.Configs.IdMappingMethod)
	if err != nil {
		return "", "", err
	}

	confidentialID, err := studyUtils.ProfileIDtoParticipantID(pID, globalSecret, study.SecretKey, study.Configs.IdMappingMethod)
	if err != nil {
		return "", "", err
	}
	return pID, confidentialID, nil
}

func saveReports(instanceID string, studyKey string, reports map[string]studyTypes.Report, withResponseID string) {
	// save reports
	for _, report := range reports {
		report.ResponseID = withResponseID
		err := studyDBService.SaveReport(instanceID, studyKey, report)
		if err != nil {
			slog.Error("Error saving report", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", report.ParticipantID), slog.String("error", err.Error()))
		} else {
			slog.Debug("Report with key '%s' for participant %s saved.", report.Key, report.ParticipantID)
		}
	}
}
