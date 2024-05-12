package study

import (
	"errors"
	"log/slog"
	"time"

	studydb "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
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

func OnEnterStudy(instanceID string, studyKey string, profileID string) (result []studyTypes.AssignedSurvey, err error) {
	study, err := studyDBService.GetStudy(instanceID, studyKey)
	if err != nil {
		slog.Error("Error getting study", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	if study.Status != studyTypes.STUDY_STATUS_ACTIVE {
		slog.Error("Study is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		err = errors.New("study is not active")
		return
	}

	participantID, confidentialID, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	// To improve privace, we reduce resolution of the timestamp to the day
	noon := time.Now().Truncate(24 * time.Hour).Add(12 * time.Hour).Unix()

	isNewParticipant := true

	// if participant exists, reuse it
	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err == nil {
		// participant exists
		slog.Debug("Participant exists", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		pState.StudyStatus = studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE
		isNewParticipant = false
	}

	if isNewParticipant {
		pState = studyTypes.Participant{
			ParticipantID: participantID,
			EnteredAt:     noon,
			StudyStatus:   studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE,
		}
	}

	if isNewParticipant {
		// save particicpant id profile lookup
		if err = studyDBService.AddConfidentialIDMapEntry(instanceID, confidentialID, profileID, studyKey); err != nil {
			slog.Error("Error saving participant ID profile lookup", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
			return
		}
	}

	currentEvent := studyengine.StudyEvent{
		Type:                                  studyengine.STUDY_EVENT_TYPE_ENTER,
		InstanceID:                            instanceID,
		StudyKey:                              studyKey,
		ParticipantIDForConfidentialResponses: confidentialID,
	}
	actionResult, err := getAndPerformStudyRules(instanceID, studyKey, pState, currentEvent)
	if err != nil {
		slog.Error("Error getting and performing study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// save participant state
	pState, err = studyDBService.SaveParticipantState(instanceID, studyKey, pState)
	if err != nil {
		slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// save reports
	saveReports(
		instanceID,
		studyKey,
		actionResult.ReportsToCreate,
		studyengine.STUDY_EVENT_TYPE_ENTER,
	)

	result = pState.AssignedSurveys
	return
}

func OnCustomStudyEvent(instanceID string, studyKey string, profileID string, eventKey string, payload map[string]interface{}) (result []studyTypes.AssignedSurvey, err error) {
	study, err := studyDBService.GetStudy(instanceID, studyKey)
	if err != nil {
		slog.Error("Error getting study", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	if study.Status != studyTypes.STUDY_STATUS_ACTIVE {
		slog.Error("Study is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		err = errors.New("study is not active")
		return
	}

	participantID, confidentialID, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		slog.Error("Error getting participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	currentEvent := studyengine.StudyEvent{
		Type:                                  studyengine.STUDY_EVENT_TYPE_CUSTOM,
		InstanceID:                            instanceID,
		StudyKey:                              studyKey,
		ParticipantIDForConfidentialResponses: confidentialID,
		EventKey:                              eventKey,
		Payload:                               payload,
	}

	actionResult, err := getAndPerformStudyRules(instanceID, studyKey, pState, currentEvent)
	if err != nil {
		slog.Error("Error getting and performing study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// save participant state
	pState, err = studyDBService.SaveParticipantState(instanceID, studyKey, pState)
	if err != nil {
		slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// save reports
	saveReports(
		instanceID,
		studyKey,
		actionResult.ReportsToCreate,
		studyengine.STUDY_EVENT_TYPE_CUSTOM,
	)

	result = pState.AssignedSurveys
	return
}

func OnLeaveStudy() {}

func OnProfileDeleted(instanceID, profileID string) {
	studies, err := studyDBService.GetStudies(instanceID, "", true)
	if err != nil {
		slog.Error("Error getting studies by status", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		return
	}

	for _, study := range studies {
		studyKey := study.Key
		slog.Info("Processing study", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))

		participantID, confidentialID, err := ComputeParticipantIDs(study, profileID)
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

func EvalCustomExpressionForParticipant(instanceID, studyKey, participantID string, expression studyTypes.Expression) (val interface{}, err error) {
	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		return nil, err
	}

	evalCtx := studyengine.EvalContext{
		Event: studyengine.StudyEvent{
			InstanceID: instanceID,
			StudyKey:   studyKey,
			Type:       studyengine.STUDY_EVENT_TYPE_CUSTOM,
		},
		ParticipantState: pState,
	}

	return studyengine.ExpressionEval(expression, evalCtx)
}

func ComputeParticipantIDs(study studyTypes.Study, profileID string) (string, string, error) {
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

func ComputeConfidentialIDForParticipant(study studyTypes.Study, participantID string) (string, error) {
	confidentialID, err := studyUtils.ProfileIDtoParticipantID(participantID, globalSecret, study.SecretKey, study.Configs.IdMappingMethod)
	if err != nil {
		return "", err
	}
	return confidentialID, nil
}
