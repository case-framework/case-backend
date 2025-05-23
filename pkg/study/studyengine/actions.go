package studyengine

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ActionEval(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	if event.Type == STUDY_EVENT_TYPE_SUBMIT {
		oldState, err = updateLastSubmissionForSurvey(oldState, event)
		if err != nil {
			return oldState, err
		}
	}

	switch action.Name {
	case "IF":
		newState, err = ifAction(action, oldState, event)
	case "DO":
		newState, err = doAction(action, oldState, event)
	case "IFTHEN":
		newState, err = ifThenAction(action, oldState, event)
	case "UPDATE_STUDY_STATUS":
		newState, err = updateStudyStatusAction(action, oldState, event)
	case "START_NEW_STUDY_SESSION":
		newState, err = startNewStudySession(action, oldState)
	case "UPDATE_FLAG":
		newState, err = updateFlagAction(action, oldState, event)
	case "REMOVE_FLAG":
		newState, err = removeFlagAction(action, oldState, event)
	case "SET_LINKING_CODE":
		newState, err = setLinkingCodeAction(action, oldState, event)
	case "DELETE_LINKING_CODE":
		newState, err = deleteLinkingCodeAction(action, oldState, event)
	case "ADD_NEW_SURVEY":
		newState, err = addNewSurveyAction(action, oldState, event)
	case "REMOVE_ALL_SURVEYS":
		newState, err = removeAllSurveys(action, oldState)
	case "REMOVE_SURVEY_BY_KEY":
		newState, err = removeSurveyByKey(action, oldState, event)
	case "REMOVE_SURVEYS_BY_KEY":
		newState, err = removeSurveysByKey(action, oldState, event)
	case "ADD_MESSAGE":
		newState, err = addMessage(action, oldState, event)
	case "REMOVE_ALL_MESSAGES":
		newState, err = removeAllMessages(oldState)
	case "REMOVE_MESSAGES_BY_TYPE":
		newState, err = removeMessagesByType(action, oldState, event)
	case "NOTIFY_RESEARCHER":
		newState, err = notifyResearcher(action, oldState, event)
	case "INIT_REPORT":
		newState, err = initReport(action, oldState, event)
	case "UPDATE_REPORT_DATA":
		newState, err = updateReportData(action, oldState, event)
	case "REMOVE_REPORT_DATA":
		newState, err = removeReportData(action, oldState, event)
	case "CANCEL_REPORT":
		newState, err = cancelReport(action, oldState, event)
	case "REMOVE_CONFIDENTIAL_RESPONSE_BY_KEY":
		newState, err = removeConfidentialResponseByKey(action, oldState, event)
	case "REMOVE_ALL_CONFIDENTIAL_RESPONSES":
		newState, err = removeAllConfidentialResponses(action, oldState, event)
	case "EXTERNAL_EVENT_HANDLER":
		newState, err = externalEventHandler(action, oldState, event)
	case "REMOVE_STUDY_CODE":
		newState, err = removeStudyCode(action, oldState, event)
	case "DRAW_STUDY_CODE_AS_LINKING_CODE":
		newState, err = drawStudyCodeAsLinkingCode(action, oldState, event)
	default:
		newState = oldState
		err = errors.New("action name not known")
	}
	if err != nil {
		slog.Debug("error when running action: ", slog.String("action", action.Name), slog.String("error", err.Error()))
	}
	return
}

func updateLastSubmissionForSurvey(oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if event.Response.Key == "" {
		return newState, errors.New("no response key found")
	}
	if newState.PState.LastSubmissions == nil {
		newState.PState.LastSubmissions = map[string]int64{}
	}

	if event.Response.ArrivedAt == 0 {
		event.Response.ArrivedAt = Now().Unix()
	}
	newState.PState.LastSubmissions[event.Response.Key] = event.Response.ArrivedAt
	return
}

func checkCondition(condition studyTypes.ExpressionArg, EvalContext EvalContext) bool {
	if !condition.IsExpression() {
		return condition.Num != 0
	}
	val, err := ExpressionEval(*condition.Exp, EvalContext)
	bVal, ok := val.(bool)
	return bVal && ok && err == nil
}

// ifAction is used to conditionally perform actions
func ifAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 2 {
		return newState, errors.New("ifAction must have at least two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	var task studyTypes.ExpressionArg
	if checkCondition(action.Data[0], EvalContext) {
		task = action.Data[1]
	} else if len(action.Data) == 3 {
		task = action.Data[2]
	}

	if task.IsExpression() {
		newState, err = ActionEval(*task.Exp, newState, event)
		if err != nil {
			return newState, err
		}
	}
	return
}

// doAction to perform a list of actions
func doAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	for _, action := range action.Data {
		if action.IsExpression() {
			newState, err = ActionEval(*action.Exp, newState, event)
			if err != nil {
				return newState, err
			}
		}
	}
	return
}

// ifThenAction is used to conditionally perform a sequence of actions
func ifThenAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 1 {
		return newState, errors.New("ifThenAction must have at least one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	if !checkCondition(action.Data[0], EvalContext) {
		return
	}
	for _, actionArg := range action.Data[1:] {
		if actionArg.IsExpression() {
			newState, err = ActionEval(*actionArg.Exp, newState, event)
			if err != nil {
				slog.Debug("error during action", slog.String("action", actionArg.Exp.Name), slog.String("error", err.Error()))
			}
		}
	}
	return
}

// updateStudyStatusAction is used to update if user is active in the study
func updateStudyStatusAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("updateStudyStatusAction must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	status, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse argument")
	}

	newState.PState.StudyStatus = status
	return
}

// startNewStudySession is used to generate a new study session ID
func startNewStudySession(action studyTypes.Expression, oldState ActionData) (newState ActionData, err error) {
	newState = oldState

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		slog.Debug("error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
	}

	newState.PState.CurrentStudySession = strconv.FormatInt(time.Now().Unix(), 16) + hex.EncodeToString(bytes)
	return
}

// updateFlagAction is used to update one of the string flags from the participant state
func updateFlagAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("updateFlagAction must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	v, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	key, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse flag key")
	}

	value := ""
	switch flagVal := v.(type) {
	case string:
		value = flagVal
	case float64:
		value = fmt.Sprintf("%f", flagVal)
	case bool:
		value = fmt.Sprintf("%t", flagVal)
	}

	if newState.PState.Flags == nil {
		newState.PState.Flags = map[string]string{}
	} else {
		newState.PState.Flags = make(map[string]string)
		for k, v := range oldState.PState.Flags {
			newState.PState.Flags[k] = v
		}
	}
	newState.PState.Flags[key] = value
	return
}

// removeFlagAction is used to update one of the string flags from the participant state
func removeFlagAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeFlagAction must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	key, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse key")
	}

	if newState.PState.Flags != nil {
		newState.PState.Flags = make(map[string]string)
		for k, v := range oldState.PState.Flags {
			newState.PState.Flags[k] = v
		}
	}

	delete(newState.PState.Flags, key)
	return
}

func setLinkingCodeAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("setLinkingCodeAction must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	key, ok := k.(string)
	if !ok {
		return newState, errors.New("could not parse key")
	}

	v, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}
	value, ok := v.(string)
	if !ok {
		return newState, errors.New("could not parse value")
	}

	if newState.PState.LinkingCodes == nil {
		newState.PState.LinkingCodes = map[string]string{}
	} else {
		newState.PState.LinkingCodes = make(map[string]string)
		for k, v := range oldState.PState.LinkingCodes {
			newState.PState.LinkingCodes[k] = v
		}
	}
	newState.PState.LinkingCodes[key] = value
	return
}

func deleteLinkingCodeAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 1 {
		// Delete all linking codes
		newState.PState.LinkingCodes = map[string]string{}
	} else {

		EvalContext := EvalContext{
			Event:            event,
			ParticipantState: newState.PState,
		}
		k, err := EvalContext.ExpressionArgResolver(action.Data[0])
		if err != nil {
			return newState, err
		}

		key, ok := k.(string)
		if !ok {
			return newState, errors.New("could not parse key")
		}
		if newState.PState.LinkingCodes != nil {
			newState.PState.LinkingCodes = make(map[string]string)
			for k, v := range oldState.PState.LinkingCodes {
				if k == key {
					continue
				}
				newState.PState.LinkingCodes[k] = v
			}
		}
	}
	return newState, nil
}

// addNewSurveyAction appends a new AssignedSurvey for the participant state
func addNewSurveyAction(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 4 {
		return newState, errors.New("addNewSurveyAction must have exactly four arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	start, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}
	end, err := EvalContext.ExpressionArgResolver(action.Data[2])
	if err != nil {
		return newState, err
	}
	c, err := EvalContext.ExpressionArgResolver(action.Data[3])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)
	validFrom, _ := start.(float64)
	validUntil, _ := end.(float64)
	category, ok4 := c.(string)

	if !ok1 || !ok4 {
		return newState, errors.New("could not parse arguments")
	}

	newSurvey := studyTypes.AssignedSurvey{
		SurveyKey:  surveyKey,
		ValidFrom:  int64(validFrom),
		ValidUntil: int64(validUntil),
		Category:   category,
	}
	newState.PState.AssignedSurveys = make([]studyTypes.AssignedSurvey, len(oldState.PState.AssignedSurveys))
	copy(newState.PState.AssignedSurveys, oldState.PState.AssignedSurveys)

	newState.PState.AssignedSurveys = append(newState.PState.AssignedSurveys, newSurvey)
	return
}

// removeAllSurveys clear the assigned survey list
func removeAllSurveys(action studyTypes.Expression, oldState ActionData) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) > 0 {
		return newState, errors.New("removeAllSurveys must not have arguments")
	}

	newState.PState.AssignedSurveys = []studyTypes.AssignedSurvey{}
	return
}

// removeSurveyByKey removes the first or last occurence of a survey
func removeSurveyByKey(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeSurveyByKey must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	pos, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)
	position, ok2 := pos.(string)

	if !ok1 || !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	as := []studyTypes.AssignedSurvey{}
	switch position {
	case "first":
		found := false
		for _, surv := range newState.PState.AssignedSurveys {
			if surv.SurveyKey == surveyKey {
				if !found {
					found = true
					continue
				}
			}
			as = append(as, surv)
		}
	case "last":
		ind := -1
		for i, surv := range newState.PState.AssignedSurveys {
			if surv.SurveyKey == surveyKey {
				ind = i
			}
		}
		if ind < 0 {
			as = newState.PState.AssignedSurveys
		} else {
			as = append(newState.PState.AssignedSurveys[:ind], newState.PState.AssignedSurveys[ind+1:]...)
		}

	default:
		return newState, errors.New("position not known")
	}
	newState.PState.AssignedSurveys = as
	return
}

// removeSurveysByKey removes all the surveys with a specific key
func removeSurveysByKey(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeSurveysByKey must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	surveyKey, ok1 := k.(string)

	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	as := []studyTypes.AssignedSurvey{}
	for _, surv := range newState.PState.AssignedSurveys {
		if surv.SurveyKey != surveyKey {
			as = append(as, surv)
		}
	}
	newState.PState.AssignedSurveys = as
	return
}

// addMessage
func addMessage(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("addMessage must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	arg1, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}
	arg2, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	messageType, ok1 := arg1.(string)
	timestamp, ok2 := arg2.(float64)

	if !ok1 || !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	newMessage := studyTypes.ParticipantMessage{
		ID:           primitive.NewObjectID().Hex(),
		Type:         messageType,
		ScheduledFor: int64(timestamp),
	}
	newState.PState.Messages = make([]studyTypes.ParticipantMessage, len(oldState.PState.Messages))
	copy(newState.PState.Messages, oldState.PState.Messages)

	newState.PState.Messages = append(newState.PState.Messages, newMessage)
	return
}

// removeAllMessages
func removeAllMessages(oldState ActionData) (newState ActionData, err error) {
	newState = oldState

	newState.PState.Messages = []studyTypes.ParticipantMessage{}
	return
}

// removeSurveysByKey removes all the surveys with a specific key
func removeMessagesByType(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeMessagesByType must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	messageType, ok1 := k.(string)

	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	messages := []studyTypes.ParticipantMessage{}
	for _, msg := range newState.PState.Messages {
		if msg.Type != messageType {
			messages = append(messages, msg)
		}
	}
	newState.PState.Messages = messages
	return
}

// notifyResearcher can save a specific message with a payload, that should be sent out to the researcher
func notifyResearcher(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 1 {
		return newState, errors.New("notifyResearcher must have at least one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	messageType, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	payload := map[string]string{}

	for i := 1; i < len(action.Data)-1; i = i + 2 {
		k, err := EvalContext.ExpressionArgResolver(action.Data[i])
		if err != nil {
			return newState, err
		}
		v, err := EvalContext.ExpressionArgResolver(action.Data[i+1])
		if err != nil {
			return newState, err
		}

		key, ok := k.(string)
		if !ok {
			return newState, errors.New("could not parse key")
		}
		value, ok := v.(string)
		if !ok {
			return newState, errors.New("could not parse value")
		}

		payload[key] = value
	}

	message := studyTypes.StudyMessage{
		Type:          messageType,
		ParticipantID: oldState.PState.ParticipantID,
		Payload:       payload,
	}

	err = CurrentStudyEngine.studyDBService.SaveResearcherMessage(event.InstanceID, event.StudyKey, message)
	if err != nil {
		slog.Error("unexpected error when saving researcher message", slog.String("error", err.Error()))
	}
	return
}

// init one empty report for the current event - if report already existing, reset report to empty report
func initReport(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("initReport must have exactly one argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	newState.ReportsToCreate[reportKey] = studyTypes.Report{
		Key:           reportKey,
		ParticipantID: oldState.PState.ParticipantID,
		Timestamp:     Now().Truncate(time.Minute).Unix(),
	}
	return
}

// update one data entry in the report with the key, if report was not initialised, init one directly
func updateReportData(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) < 3 {
		return newState, errors.New("updateReportData must have at least 3 arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		slog.Error("unexpected error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	// If report not initialized yet, init report:
	report, hasKey := newState.ReportsToCreate[reportKey]
	if !hasKey {
		report = studyTypes.Report{
			Key:           reportKey,
			ParticipantID: oldState.PState.ParticipantID,
			Timestamp:     Now().Truncate(time.Minute).Unix(),
		}
	}

	// Get attribute Key
	a, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		slog.Debug("error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
		return newState, err
	}
	attributeKey, ok1 := a.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	// Get value
	v, err := EvalContext.ExpressionArgResolver(action.Data[2])
	if err != nil {
		slog.Debug("error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
		return newState, err
	}

	dType := ""
	if len(action.Data) > 3 {
		// Set dtype
		d, err := EvalContext.ExpressionArgResolver(action.Data[3])
		if err != nil {
			return newState, err
		}
		dtype, ok1 := d.(string)
		if !ok1 {
			return newState, errors.New("could not parse arguments")
		}
		dType = dtype
	}

	value := ""
	switch flagVal := v.(type) {
	case string:
		value = flagVal
	case float64:
		if dType == "int" {
			value = fmt.Sprintf("%d", int(flagVal))
		} else {
			value = fmt.Sprintf("%f", flagVal)
		}
	case bool:
		value = fmt.Sprintf("%t", flagVal)
	}

	newData := studyTypes.ReportData{
		Key:   attributeKey,
		Value: value,
		Dtype: dType,
	}

	index := -1
	for i, d := range report.Data {
		if d.Key == attributeKey {
			index = i
			break
		}
	}

	if index < 0 {
		report.Data = append(report.Data, newData)
	} else {
		report.Data[index] = newData
	}

	newState.ReportsToCreate[reportKey] = report
	return
}

// remove one data entry in the report with the key
func removeReportData(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 2 {
		return newState, errors.New("removeReportData must have exactly two arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	// If report not initialized yet, init report:
	report, hasKey := newState.ReportsToCreate[reportKey]
	if !hasKey {
		// nothing to do
		return newState, nil
	}

	// Get attribute Key
	a, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}
	attributeKey, ok1 := a.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	index := -1
	for i, d := range report.Data {
		if d.Key == attributeKey {
			index = i
			break
		}
	}

	if index > -1 {
		report.Data = append(report.Data[:index], report.Data[index+1:]...)
	}

	newState.ReportsToCreate[reportKey] = report
	return
}

// remove the report from this event
func cancelReport(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("updateReportData must have exactly 1 argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	reportKey, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	_, hasKey := newState.ReportsToCreate[reportKey]
	if hasKey {
		delete(newState.ReportsToCreate, reportKey)
	}
	return
}

// delete confidential responses for this participant for a particular key only
func removeConfidentialResponseByKey(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	if len(action.Data) != 1 {
		return newState, errors.New("removeConfidentialResponseByKey must have exactly 1 argument")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	key, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	_, err = CurrentStudyEngine.studyDBService.DeleteConfidentialResponses(event.InstanceID, event.StudyKey, event.ParticipantIDForConfidentialResponses, key)
	if err != nil {
		slog.Error("unexpected error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
	}
	return
}

// delete confidential responses for this participant
func removeAllConfidentialResponses(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState
	_, err = CurrentStudyEngine.studyDBService.DeleteConfidentialResponses(event.InstanceID, event.StudyKey, event.ParticipantIDForConfidentialResponses, "")
	if err != nil {
		slog.Error("unexpected error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
	}
	return
}

// call external service to handle event
func externalEventHandler(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState

	if len(action.Data) < 1 {
		msg := "externalEventHandler must have at least 1 argument"
		slog.Error("unexpected error during action", slog.String("action", action.Name), slog.String("error", msg))
		return newState, errors.New(msg)
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	k, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	serviceName, ok1 := k.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	serviceConfig, err := getExternalServicesConfigByName(serviceName)
	if err != nil {
		slog.Error("unexpected error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
		return newState, err
	}

	pathname := ""
	if len(action.Data) > 1 {
		arg1, err := EvalContext.ExpressionArgResolver(action.Data[1])
		if err != nil {
			return newState, err
		}

		route := arg1.(string)
		route = strings.TrimPrefix(route, "/")
		pathname = route
	}

	var mTLSConfig *apihelpers.CertificatePaths
	if serviceConfig.MutualTLSConfig != nil {
		mTLSConfig = &apihelpers.CertificatePaths{
			CACertPath:     serviceConfig.MutualTLSConfig.CAFile,
			ServerCertPath: serviceConfig.MutualTLSConfig.CertFile,
			ServerKeyPath:  serviceConfig.MutualTLSConfig.KeyFile,
		}
	}

	httpClient := httpclient.ClientConfig{
		RootURL:                   serviceConfig.URL,
		APIKey:                    serviceConfig.APIKey,
		Timeout:                   time.Duration(serviceConfig.Timeout) * time.Second,
		MutualTLSCertificatePaths: mTLSConfig,
	}

	payload := ExternalEventPayload{
		ParticipantState: newState.PState,
		EventType:        event.Type,
		StudyKey:         event.StudyKey,
		InstanceID:       event.InstanceID,
		Response:         event.Response,
		EventKey:         event.EventKey,
		Payload:          event.Payload,
	}

	response, err := httpClient.RunHTTPcall(pathname, payload)
	if err != nil {
		slog.Debug("unexpected error with external event handler", slog.String("action", action.Name), slog.String("serviceName", serviceName), slog.String("error", err.Error()))
		return newState, err
	}

	// if relevant, update participant state:
	pState, hasKey := response["pState"]
	if hasKey {
		newState.PState = pState.(studyTypes.Participant)
		slog.Debug("received new participant state from external service")
	}

	// collect reports if any:
	reportsToCreate, hasKey := response["reportsToCreate"]
	if hasKey {
		reportsToCreate := reportsToCreate.(map[string]studyTypes.Report)
		for key, value := range reportsToCreate {
			newState.ReportsToCreate[key] = value
		}
		slog.Debug("received new report list from external service")
	}
	return
}

func removeStudyCode(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState

	if len(action.Data) != 2 {
		return newState, errors.New("invalid number of arguments")
	}
	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	arg1, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	listKey, ok1 := arg1.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	arg2, err := EvalContext.ExpressionArgResolver(action.Data[1])
	if err != nil {
		return newState, err
	}

	code, ok2 := arg2.(string)
	if !ok2 {
		return newState, errors.New("could not parse arguments")
	}

	err = CurrentStudyEngine.studyDBService.DeleteStudyCodeListEntry(event.InstanceID, event.StudyKey, listKey, code)
	if err != nil {
		slog.Error("unexpected error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
	}
	return
}

func drawStudyCodeAsLinkingCode(action studyTypes.Expression, oldState ActionData, event StudyEvent) (newState ActionData, err error) {
	newState = oldState

	if len(action.Data) < 1 {
		return newState, errors.New("DRAW_STUDY_CODE_AS_LINKING_CODE must have at least one argument")
	}

	EvalContext := EvalContext{
		Event:            event,
		ParticipantState: newState.PState,
	}
	arg1, err := EvalContext.ExpressionArgResolver(action.Data[0])
	if err != nil {
		return newState, err
	}

	listKey, ok1 := arg1.(string)
	if !ok1 {
		return newState, errors.New("could not parse arguments")
	}

	linkingCodeKey := listKey
	if len(action.Data) > 1 {
		arg2, err := EvalContext.ExpressionArgResolver(action.Data[1])

		if err != nil {
			return newState, err
		}
		var ok2 bool
		linkingCodeKey, ok2 = arg2.(string)
		if !ok2 {
			return newState, errors.New("could not parse arguments")
		}
	}

	if newState.PState.LinkingCodes == nil {
		newState.PState.LinkingCodes = map[string]string{}
	} else {
		newState.PState.LinkingCodes = make(map[string]string)
		for k, v := range oldState.PState.LinkingCodes {
			newState.PState.LinkingCodes[k] = v
		}
	}

	// draw code
	code, err := CurrentStudyEngine.studyDBService.DrawStudyCode(event.InstanceID, event.StudyKey, listKey)
	if err != nil {
		slog.Error("unexpected error during action", slog.String("action", action.Name), slog.String("error", err.Error()))
		return newState, err
	}

	// if code emptry, remove linking code
	if code == "" {
		slog.Debug("linking code is empty, removing")
		delete(newState.PState.LinkingCodes, linkingCodeKey)
	} else {
		newState.PState.LinkingCodes[linkingCodeKey] = code
	}

	return
}
