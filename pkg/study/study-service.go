package study

import (
	"errors"
	"log/slog"
	"time"

	studydb "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	studyUtils "github.com/case-framework/case-backend/pkg/study/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	studyDBService *studydb.StudyDBService
	globalSecret   string
)

const (
	TEMPORARY_PARTICIPANT_TAKEOVER_PERIOD = 24 * 60 * 60 // seconds - after this period, the temporary participant is considered to be inactive and cannot be used anymore
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
	pState, err = studyDBService.SaveParticipantState(instanceID, studyKey, actionResult.PState)
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

func OnRegisterTempParticipant(instanceID string, studyKey string) (pState *studyTypes.Participant, err error) {
	study, err := studyDBService.GetStudy(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	if study.Status != studyTypes.STUDY_STATUS_ACTIVE {
		slog.Error("study is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		err = errors.New("study is not active")
		return
	}

	tempProfileID := primitive.NewObjectID().Hex()
	participantID, _, err := ComputeParticipantIDs(study, tempProfileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	pState = &studyTypes.Participant{
		ParticipantID: participantID,
		StudyStatus:   studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY,
		EnteredAt:     time.Now().Unix(),
	}

	currentEvent := studyengine.StudyEvent{
		Type:       studyengine.STUDY_EVENT_TYPE_ENTER,
		InstanceID: instanceID,
		StudyKey:   studyKey,
	}

	actionResult, err := getAndPerformStudyRules(instanceID, studyKey, *pState, currentEvent)
	if err != nil {
		slog.Error("Error getting and performing study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// save participant state
	_, err = studyDBService.SaveParticipantState(instanceID, studyKey, actionResult.PState)
	if err != nil {
		slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	saveReports(instanceID, studyKey, actionResult.ReportsToCreate, studyengine.STUDY_EVENT_TYPE_ENTER)
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
	pState, err = studyDBService.SaveParticipantState(instanceID, studyKey, actionResult.PState)
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

func OnMergeTempParticipant(instanceID string, studyKey string, profileID string, temporaryParticipantID string) (result []studyTypes.AssignedSurvey, err error) {
	study, err := studyDBService.GetStudy(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	if study.Status != studyTypes.STUDY_STATUS_ACTIVE {
		slog.Error("study is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		err = errors.New("study is not active")
		return
	}

	tempParticipantState, err := studyDBService.GetParticipantByID(instanceID, studyKey, temporaryParticipantID)
	if err != nil {
		slog.Error("error getting temporary participant", slog.String("error", err.Error()))
		return
	}

	if tempParticipantState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY {
		slog.Error("temporary participant is not temporary", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", temporaryParticipantID))
		err = errors.New("temporary participant is not temporary")
		return
	}

	if tempParticipantState.EnteredAt+TEMPORARY_PARTICIPANT_TAKEOVER_PERIOD < time.Now().Unix() {
		// This is to prevent takeover of temporary participants by brute force trial
		time.Sleep(10 * time.Second)
		slog.Error("temporary participant is too old", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", temporaryParticipantID))
		err = errors.New("temporary participant is too old")
		return
	}

	participantID, confidentialID, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		slog.Info("participant not found, creating new one", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		pState = studyTypes.Participant{
			ParticipantID: participantID,
			StudyStatus:   studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE,
			EnteredAt:     time.Now().Unix(),
		}

		// save lookup for participant ID
		err = studyDBService.AddConfidentialIDMapEntry(instanceID, confidentialID, profileID, studyKey)
		if err != nil {
			slog.Error("Error saving participant ID lookup", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
			return
		}
	}

	// Merge participant states
	currentEvent := studyengine.StudyEvent{
		InstanceID:                            instanceID,
		StudyKey:                              studyKey,
		Type:                                  studyengine.STUDY_EVENT_TYPE_MERGE,
		MergeWithParticipant:                  tempParticipantState,
		ParticipantIDForConfidentialResponses: confidentialID,
	}

	actionResult, err := getAndPerformStudyRules(instanceID, studyKey, pState, currentEvent)
	if err != nil {
		slog.Error("Error getting and performing study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// save participant state
	pState, err = studyDBService.SaveParticipantState(instanceID, studyKey, actionResult.PState)
	if err != nil {
		slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// delete temporary participant
	err = studyDBService.DeleteParticipantByID(instanceID, studyKey, temporaryParticipantID)
	if err != nil {
		slog.Error("Error deleting temporary participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// update participant ID to all response object
	count, err := studyDBService.UpdateParticipantIDonResponses(instanceID, studyKey, temporaryParticipantID, participantID)
	if err != nil {
		slog.Error("Error updating participant ID on responses", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
	} else {
		slog.Debug("updated responses for participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.Int64("count", count))
	}

	// update participant ID to all history object
	count, err = studyDBService.UpdateParticipantIDonReports(instanceID, studyKey, temporaryParticipantID, participantID)
	if err != nil {
		slog.Error("Error updating participant ID on reports", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
	} else {
		slog.Debug("updated reports for participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.Int64("count", count))
	}

	// update participant ID to all confidential responses
	oldConfidentialID, err := ComputeConfidentialIDForParticipant(study, temporaryParticipantID)
	if err != nil {
		slog.Error("Error computing confidential ID", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}
	count, err = studyDBService.UpdateParticipantIDonConfidentialResponses(instanceID, studyKey, oldConfidentialID, participantID)
	if err != nil {
		slog.Error("Error updating participant ID on confidential responses", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
	} else {
		slog.Debug("updated confidential responses for participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.Int64("count", count))
	}

	result = pState.AssignedSurveys
	return
}

func OnSubmitResponseForTempParticipant(instanceID string, studyKey string, participantID string, response studyTypes.SurveyResponse) (result []studyTypes.AssignedSurvey, err error) {
	study, err := studyDBService.GetStudy(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	if study.Status != studyTypes.STUDY_STATUS_ACTIVE {
		slog.Error("study is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		err = errors.New("study is not active")
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		slog.Error("error getting participant", slog.String("error", err.Error()))
		return
	}

	if pState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY {
		slog.Error("participant is not temporary", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		err = errors.New("participant is not temporary")
		return
	}

	if pState.EnteredAt+TEMPORARY_PARTICIPANT_TAKEOVER_PERIOD < time.Now().Unix() {
		// This is to prevent takeover of temporary participants by brute force trial
		time.Sleep(10 * time.Second)
		slog.Error("temporary participant is too old", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		err = errors.New("temporary participant is too old")
		return
	}

	confidentialID, err := ComputeConfidentialIDForParticipant(study, participantID)
	if err != nil {
		slog.Error("Error computing confidential ID", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	currentEvent := studyengine.StudyEvent{
		Type:                                  studyengine.STUDY_EVENT_TYPE_SUBMIT,
		InstanceID:                            instanceID,
		StudyKey:                              studyKey,
		Response:                              response,
		ParticipantIDForConfidentialResponses: confidentialID,
	}
	actionResult, err := getAndPerformStudyRules(instanceID, studyKey, pState, currentEvent)
	if err != nil {
		slog.Error("Error getting and performing study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	// save participant state
	_, err = studyDBService.SaveParticipantState(instanceID, studyKey, actionResult.PState)
	if err != nil {
		slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	responseId, err := saveResponses(instanceID, studyKey, response, pState, confidentialID)
	if err != nil {
		slog.Error("Error saving responses", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	saveReports(instanceID, studyKey, actionResult.ReportsToCreate, responseId)

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
