package study

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"time"

	studydb "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	"github.com/case-framework/case-backend/pkg/study/types"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	studyUtils "github.com/case-framework/case-backend/pkg/study/utils"
	"go.mongodb.org/mongo-driver/bson"
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
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
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

		if pState.StudyStatus == studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE {
			slog.Debug("Participant is already active, do not run study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
			return pState.AssignedSurveys, nil
		}
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
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
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
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
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
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
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
	count, err = studyDBService.UpdateParticipantIDonConfidentialResponses(instanceID, studyKey, oldConfidentialID, confidentialID)
	if err != nil {
		slog.Error("Error updating participant ID on confidential responses", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
	} else {
		slog.Debug("updated confidential responses for participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.Int64("count", count))
	}

	// delete temporary participant
	err = studyDBService.DeleteParticipantByID(instanceID, studyKey, temporaryParticipantID)
	if err != nil {
		slog.Error("Error deleting temporary participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	err = nil
	result = pState.AssignedSurveys
	return
}

func OnSubmitResponse(instanceID string, studyKey string, profileID string, response studyTypes.SurveyResponse) (result []studyTypes.AssignedSurvey, err error) {
	response.ArrivedAt = time.Now().Unix()

	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	participantID, confidentialID, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		slog.Error("error getting participant state", slog.String("error", err.Error()))
		return
	}

	if pState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE {
		slog.Error("participant is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		err = errors.New("participant is not active")
		return
	}

	currentEvent := studyengine.StudyEvent{
		Type:                                  studyengine.STUDY_EVENT_TYPE_SUBMIT,
		InstanceID:                            instanceID,
		StudyKey:                              studyKey,
		ParticipantIDForConfidentialResponses: confidentialID,
		Response:                              response,
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

	result = make([]studyTypes.AssignedSurvey, len(actionResult.PState.AssignedSurveys))
	for i, survey := range actionResult.PState.AssignedSurveys {
		result[i] = survey
		result[i].ProfileID = profileID
		result[i].StudyKey = studyKey
	}
	return
}

func OnSubmitResponseForTempParticipant(instanceID string, studyKey string, participantID string, response studyTypes.SurveyResponse) (result []studyTypes.AssignedSurvey, err error) {
	response.ArrivedAt = time.Now().Unix()

	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
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

type RunStudyActionProgressFn func(totalCount int64, processedCount int64)

type RunStudyActionReq struct {
	InstanceID           string
	StudyKey             string
	OnlyForParticipantID string
	Rules                []types.Expression
	OnProgressFn         RunStudyActionProgressFn
}

type RunStudyActionResult struct {
	ParticipantCount               int64
	ParticipantStateChangedPerRule []int64
	Duration                       int64
}

func OnRunStudyAction(req RunStudyActionReq) (*RunStudyActionResult, error) {
	if studyDBService == nil {
		return nil, errors.New("studyDBService is not initialized")
	}

	if req.InstanceID == "" || req.StudyKey == "" {
		return nil, errors.New("instanceID and studyKey are required")
	}

	filter := bson.M{
		"studyStatus": bson.M{"$nin": []string{
			studyTypes.PARTICIPANT_STUDY_STATUS_ACCOUNT_DELETED,
			studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY,
		}},
	}

	if req.OnlyForParticipantID != "" {
		filter = bson.M{
			"participantID": req.OnlyForParticipantID,
		}
	}

	study, err := studyDBService.GetStudy(req.InstanceID, req.StudyKey)
	if err != nil {
		return nil, err
	}

	count, err := studyDBService.GetParticipantCount(req.InstanceID, req.StudyKey, filter)
	if err != nil {
		return nil, err
	}

	result := &RunStudyActionResult{
		ParticipantCount:               0,
		ParticipantStateChangedPerRule: make([]int64, len(req.Rules)),
		Duration:                       0,
	}
	start := time.Now().Unix()

	if req.OnProgressFn != nil {
		req.OnProgressFn(count, 0)
	}

	err = studyDBService.FindAndExecuteOnParticipantsStates(
		context.Background(),
		req.InstanceID,
		req.StudyKey,
		filter,
		nil,
		false,
		func(dbService *studydb.StudyDBService, p studyTypes.Participant, instanceID, studyKey string, args ...interface{}) error {
			result.ParticipantCount += 1

			if req.OnProgressFn != nil {
				req.OnProgressFn(count, result.ParticipantCount)
			}

			confidentialID, err := ComputeConfidentialIDForParticipant(study, p.ParticipantID)
			if err != nil {
				slog.Error("Error computing confidential ID", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
				return err
			}

			participantData := studyengine.ActionData{
				PState:          p,
				ReportsToCreate: map[string]studyTypes.Report{},
			}

			anyChange := false

			for i, rule := range req.Rules {
				event := studyengine.StudyEvent{
					InstanceID:                            instanceID,
					StudyKey:                              studyKey,
					Type:                                  studyengine.STUDY_EVENT_TYPE_CUSTOM,
					ParticipantIDForConfidentialResponses: confidentialID,
				}

				newState, err := studyengine.ActionEval(rule, participantData, event)
				if err != nil {
					slog.Error("Error evaluating study rule", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("rule", rule.Name), slog.String("error", err.Error()))
					return err
				}

				if !reflect.DeepEqual(newState.PState, participantData.PState) {
					result.ParticipantStateChangedPerRule[i] += 1
					anyChange = true
				}
				participantData = newState
			}

			if anyChange {
				_, err = studyDBService.SaveParticipantState(instanceID, studyKey, participantData.PState)
				if err != nil {
					slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
					return err
				}
			}

			saveReports(instanceID, studyKey, participantData.ReportsToCreate, studyengine.STUDY_EVENT_TYPE_CUSTOM)

			result.Duration = time.Now().Unix() - start
			return nil
		},
	)
	if err != nil {
		slog.Error("Error executing study action", slog.String("instanceID", req.InstanceID), slog.String("studyKey", req.StudyKey), slog.String("error", err.Error()))
	}

	result.Duration = time.Now().Unix() - start

	return result, nil
}

func OnRunStudyActionForPreviousResponses(req RunStudyActionReq, surveyKeys []string, from int64, to int64) (*RunStudyActionResult, error) {
	if req.InstanceID == "" || req.StudyKey == "" {
		return nil, errors.New("instanceID and studyKey are required")
	}

	filter := bson.M{}

	if req.OnlyForParticipantID != "" {
		filter = bson.M{
			"participantID": req.OnlyForParticipantID,
		}
	}

	study, err := studyDBService.GetStudy(req.InstanceID, req.StudyKey)
	if err != nil {
		return nil, err
	}

	count, err := studyDBService.GetParticipantCount(req.InstanceID, req.StudyKey, filter)
	if err != nil {
		return nil, err
	}

	result := &RunStudyActionResult{
		ParticipantCount:               0,
		ParticipantStateChangedPerRule: make([]int64, len(req.Rules)),
		Duration:                       0,
	}
	start := time.Now().Unix()

	if req.OnProgressFn != nil {
		req.OnProgressFn(count, 0)
	}

	err = studyDBService.FindAndExecuteOnParticipantsStates(
		context.Background(),
		req.InstanceID,
		req.StudyKey,
		filter,
		nil,
		false,
		func(dbService *studydb.StudyDBService, p studyTypes.Participant, instanceID, studyKey string, args ...interface{}) error {
			result.ParticipantCount += 1

			if req.OnProgressFn != nil {
				req.OnProgressFn(count, result.ParticipantCount)
			}

			confidentialID, err := ComputeConfidentialIDForParticipant(study, p.ParticipantID)
			if err != nil {
				slog.Error("Error computing confidential ID", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
				return err
			}

			responseFilter := bson.M{
				"participantID": p.ParticipantID,
			}
			if from > 0 && to > 0 {
				responseFilter["arrivedAt"] = bson.M{"$and": []bson.M{
					{"$gt": from},
					{"$lt": to},
				}}
			} else if from > 0 {
				responseFilter["arrivedAt"] = bson.M{"$gt": from}
			} else if to > 0 {
				responseFilter["arrivedAt"] = bson.M{"$lt": to}
			}

			if len(surveyKeys) > 0 {
				responseFilter["key"] = bson.M{"$in": surveyKeys}
			}

			sort := bson.M{
				"arrivedAt": 1,
			}

			err = studyDBService.FindAndExecuteOnResponses(
				context.Background(),
				instanceID,
				studyKey,
				responseFilter,
				sort,
				false,
				func(dbService *studydb.StudyDBService, r studyTypes.SurveyResponse, instanceID, studyKey string, args ...interface{}) error {
					freshPState, err := dbService.GetParticipantByID(instanceID, studyKey, p.ParticipantID)
					if err != nil {
						slog.Error("Error getting participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
						return err
					}

					participantData := studyengine.ActionData{
						PState:          freshPState,
						ReportsToCreate: map[string]studyTypes.Report{},
					}

					for _, rule := range req.Rules {
						event := studyengine.StudyEvent{
							InstanceID:                            instanceID,
							StudyKey:                              studyKey,
							Type:                                  studyengine.STUDY_EVENT_TYPE_SUBMIT,
							ParticipantIDForConfidentialResponses: confidentialID,
							Response:                              r,
						}

						newState, err := studyengine.ActionEval(rule, participantData, event)
						if err != nil {
							slog.Error("Error evaluating study rule", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("rule", rule.Name), slog.String("error", err.Error()))
							return err
						}
						participantData = newState

					}

					for key, report := range participantData.ReportsToCreate {
						report.Timestamp = r.SubmittedAt
						participantData.ReportsToCreate[key] = report
					}
					saveReports(instanceID, studyKey, participantData.ReportsToCreate, r.ID.Hex())

					_, err = studyDBService.SaveParticipantState(instanceID, studyKey, participantData.PState)
					if err != nil {
						slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
						return err
					}

					return nil
				},
			)
			if err != nil {
				slog.Error("Error executing study action on previous responses", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
				return err
			}

			return nil
		},
	)
	if err != nil {
		slog.Error("Error executing study action", slog.String("instanceID", req.InstanceID), slog.String("studyKey", req.StudyKey), slog.String("error", err.Error()))
	}

	result.Duration = time.Now().Unix() - start

	return result, nil
}

// Run study timer event for participants
func OnStudyTimer(instanceID string, study *studyTypes.Study) {
	if study == nil {
		slog.Error("study is nil", slog.String("instanceID", instanceID))
		return
	}
	rulesObj, err := studyDBService.GetCurrentStudyRules(instanceID, study.Key)
	if err != nil {
		return
	}

	currentEvent := studyengine.StudyEvent{
		Type:       studyengine.STUDY_EVENT_TYPE_TIMER,
		InstanceID: instanceID,
		StudyKey:   study.Key,
	}

	if !hasRuleForEventType(rulesObj.Rules, currentEvent) {
		slog.Debug("no timer event rules found", slog.String("instanceID", instanceID), slog.String("studyKey", study.Key))
		return
	}

	filter := bson.M{
		"studyStatus": bson.M{"$nin": []string{
			studyTypes.PARTICIPANT_STUDY_STATUS_ACCOUNT_DELETED,
			studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY,
		}},
	}

	err = studyDBService.FindAndExecuteOnParticipantsStates(
		context.Background(),
		instanceID,
		study.Key,
		filter,
		nil,
		false,
		func(dbService *studydb.StudyDBService, p studyTypes.Participant, instanceID string, studyKey string, args ...interface{}) error {
			confidentialID, err := ComputeConfidentialIDForParticipant(*study, p.ParticipantID)
			if err != nil {
				slog.Error("Error computing confidential ID", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
				return err
			}

			currentEvent.ParticipantIDForConfidentialResponses = confidentialID

			newState := studyengine.ActionData{
				PState:          p,
				ReportsToCreate: map[string]studyTypes.Report{},
			}

			for _, rule := range rulesObj.Rules {
				newState, err = studyengine.ActionEval(rule, newState, currentEvent)
				if err != nil {
					slog.Error("Error evaluating study rule", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
					continue
				}
			}

			// save participant state
			_, err = studyDBService.SaveParticipantState(instanceID, studyKey, newState.PState)
			if err != nil {
				slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", p.ParticipantID), slog.String("error", err.Error()))
				return err
			}

			saveReports(instanceID, studyKey, newState.ReportsToCreate, studyengine.STUDY_EVENT_TYPE_TIMER)

			return nil
		},
	)
	if err != nil {
		slog.Error("Error executing study timer event", slog.String("instanceID", instanceID), slog.String("studyKey", study.Key), slog.String("error", err.Error()))
	}
}

func OnLeaveStudy(instanceID string, studyKey string, profileID string) (result []studyTypes.AssignedSurvey, err error) {
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	participantID, confidentialID, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, profileID)
	if err != nil {
		slog.Error("error getting participant state", slog.String("error", err.Error()))
		return
	}

	if pState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE {
		slog.Error("participant is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", profileID))
		err = errors.New("participant is not active")
		return
	}

	pState.StudyStatus = studyTypes.PARTICIPANT_STUDY_STATUS_EXITED

	currentEvent := studyengine.StudyEvent{
		Type:                                  studyengine.STUDY_EVENT_TYPE_LEAVE,
		InstanceID:                            instanceID,
		StudyKey:                              studyKey,
		ParticipantIDForConfidentialResponses: confidentialID,
	}

	actionResult, err := getAndPerformStudyRules(instanceID, studyKey, pState, currentEvent)
	if err != nil {
		slog.Error("Error getting and performing study rules", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	_, err = studyDBService.SaveParticipantState(instanceID, studyKey, actionResult.PState)
	if err != nil {
		slog.Error("Error saving participant state", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
		return
	}

	saveReports(instanceID, studyKey, actionResult.ReportsToCreate, studyengine.STUDY_EVENT_TYPE_LEAVE)

	_, err = studyDBService.DeleteConfidentialResponses(instanceID, studyKey, confidentialID, "")
	if err != nil {
		slog.Error("Error deleting confidential responses", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("error", err.Error()))
	}
	result = pState.AssignedSurveys
	return
}

func OnProfileDeleted(instanceID, profileID string, exitSurveyResp *studyTypes.SurveyResponse) {
	if exitSurveyResp != nil {
		exitSurveyResp.ArrivedAt = time.Now().Unix()
	}
	studies, err := studyDBService.GetStudies(instanceID, "", false)
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

		// save exit survey response even if no participant state is found, if it's a system default study
		if study.Props.SystemDefaultStudy && exitSurveyResp != nil {
			_, err := saveResponses(instanceID, study.Key, *exitSurveyResp, studyTypes.Participant{
				ParticipantID: participantID,
			}, confidentialID)
			if err != nil {
				slog.Error("Error saving responses", slog.String("instanceID", instanceID), slog.String("studyKey", study.Key), slog.String("participantID", participantID), slog.String("error", err.Error()))
				return
			}
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
