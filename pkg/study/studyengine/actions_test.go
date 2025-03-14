package studyengine

import (
	"strconv"
	"testing"
	"time"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

const (
	ASSIGNED_SURVEY_CATEGORY_PRIO   = "prio"
	ASSIGNED_SURVEY_CATEGORY_NORMAL = "normal"
)

func TestActions(t *testing.T) {
	// Override Now function for testing
	originalNow := Now
	defer func() { Now = originalNow }()
	Now = func() time.Time {
		return time.Unix(1609459200, 0) // Fixed time for testing: 2021-01-01 00:00:00 UTC
	}

	actionData := ActionData{
		PState: studyTypes.Participant{
			ParticipantID: "participant1234",
			StudyStatus:   studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE,
			Flags: map[string]string{
				"health": "test",
			},
		},
		ReportsToCreate: map[string]studyTypes.Report{},
	}
	event := StudyEvent{
		Type: "SUBMIT",
		Response: studyTypes.SurveyResponse{
			Key: "test",
		},
	}

	t.Run("with wrong action name", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "WRONG",
		}
		_, err := ActionEval(action, actionData, event)
		if err == nil {
			t.Error("should return an error")
		}
	})

	t.Run("IFTHEN", func(t *testing.T) {
		action2 := studyTypes.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testflag_cond"},
			},
		}
		action3 := studyTypes.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testflag_cond2"},
			},
		}
		action := studyTypes.Expression{
			Name: "IFTHEN",
			Data: []studyTypes.ExpressionArg{
				{DType: "num", Num: 0},
				{DType: "exp", Exp: &action2},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.PState.StudyStatus == action2.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action2.Data[0].Str, newState.PState.StudyStatus)
		}

		action = studyTypes.Expression{
			Name: "IFTHEN",
			Data: []studyTypes.ExpressionArg{
				{DType: "num", Num: 1},
				{DType: "exp", Exp: &action2},
				{DType: "exp", Exp: &action3},
			},
		}
		newState, err = ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.PState.StudyStatus != action3.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action3.Data[0].Str, newState.PState.StudyStatus)
		}
	})

	t.Run("IF", func(t *testing.T) {
		action2 := studyTypes.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testflag_cond"},
			},
		}
		action3 := studyTypes.Expression{
			Name: "UPDATE_STUDY_STATUS",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testflag_cond2"},
			},
		}
		action := studyTypes.Expression{
			Name: "IF",
			Data: []studyTypes.ExpressionArg{
				{DType: "num", Num: 0},
				{DType: "exp", Exp: &action2},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.PState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE {
			t.Errorf("error -> expected: %s, have: %s", studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE, newState.PState.StudyStatus)
		}

		action = studyTypes.Expression{
			Name: "IF",
			Data: []studyTypes.ExpressionArg{
				{DType: "num", Num: 1},
				{DType: "exp", Exp: &action2},
			},
		}
		newState, err = ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.PState.StudyStatus != action2.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action.Data[1].Str, newState.PState.StudyStatus)
		}

		action = studyTypes.Expression{
			Name: "IF",
			Data: []studyTypes.ExpressionArg{
				{DType: "num", Num: 0},
				{DType: "exp", Exp: &action2},
				{DType: "exp", Exp: &action3},
			},
		}
		newState, err = ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if newState.PState.StudyStatus != action3.Data[0].Str {
			t.Errorf("error -> expected: %s, have: %s", action3.Data[0].Str, newState.PState.StudyStatus)
		}
	})

	t.Run("UPDATE_FLAG", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "UPDATE_FLAG",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "key"},
				{DType: "str", Str: "value"},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		v, ok := newState.PState.Flags["key"]
		if !ok {
			t.Error("could not find new flag")
			return
		}
		if v != action.Data[1].Str {
			t.Errorf("updated status error -> expected: %s, have: %s", action.Data[1].Str, v)
		}
	})

	t.Run("UPDATE_FLAG with number", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "UPDATE_FLAG",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "keyNum"},
				{DType: "num", Num: 14},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		v, ok := newState.PState.Flags["keyNum"]
		if !ok {
			t.Error("could not find new flag")
			return
		}
		res, err := strconv.ParseFloat(v, 64)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if res != action.Data[1].Num {
			t.Errorf("updated status error -> expected: %s, have: %s", action.Data[1].Str, v)
		}
	})

	t.Run("REMOVE_FLAG", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "REMOVE_FLAG",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "health"},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		_, ok := newState.PState.Flags["health"]
		if ok {
			t.Error("should not find value")
			return
		}
	})

	// Survey actions:
	t.Run("ADD_NEW_SURVEY", func(t *testing.T) {
		now := time.Now().Unix()
		action := studyTypes.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.AssignedSurveys) != 1 {
			t.Errorf("updated number of surveys: %d", len(newState.PState.AssignedSurveys))
			return
		}
		if newState.PState.AssignedSurveys[0].ValidFrom != now-10 {
			t.Errorf("unexpected validFrom: have %d, exprected: %d", newState.PState.AssignedSurveys[0].ValidFrom, now-10)
		}
		if newState.PState.AssignedSurveys[0].ValidUntil != now+10 {
			t.Errorf("unexpected validFrom: have %d, exprected: %d", newState.PState.AssignedSurveys[0].ValidUntil, now+10)
		}
	})

	t.Run("REMOVE_ALL_SURVEYS", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := studyTypes.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.AssignedSurveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(newState.PState.AssignedSurveys))
			return
		}

		// REMOVE_ALL_SURVEYS
		action = studyTypes.Expression{
			Name: "REMOVE_ALL_SURVEYS",
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.PState.AssignedSurveys) > 0 {
			t.Error("should not have surveys any more")
		}
	})

	t.Run("REMOVE_SURVEY_BY_KEY last", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := studyTypes.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey1"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.PState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEY_BY_KEY
		action = studyTypes.Expression{
			Name: "REMOVE_SURVEY_BY_KEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "str", Str: "last"},
			},
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.PState.AssignedSurveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(newState.PState.AssignedSurveys))
			return
		}
		if newState.PState.AssignedSurveys[0].SurveyKey != "testSurveyKey1" {
			t.Errorf("unexpected survey key at pos 0: %s", newState.PState.AssignedSurveys[0].SurveyKey)
		}
	})

	t.Run("REMOVE_SURVEY_BY_KEY first", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := studyTypes.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey1"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.PState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEY_BY_KEY
		action = studyTypes.Expression{
			Name: "REMOVE_SURVEY_BY_KEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "str", Str: "first"},
			},
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.PState.AssignedSurveys) != 2 {
			t.Errorf("unexpected number of surveys: %d", len(newState.PState.AssignedSurveys))
		}
		if newState.PState.AssignedSurveys[0].SurveyKey != "testSurveyKey2" {
			t.Errorf("unexpected survey key at pos 0: %s", newState.PState.AssignedSurveys[0].SurveyKey)
		}
	})

	t.Run("REMOVE_SURVEYS_BY_KEY", func(t *testing.T) {
		// Add surveys first
		now := time.Now().Unix()
		action := studyTypes.Expression{
			Name: "ADD_NEW_SURVEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
				{DType: "num", Num: float64(now - 10)},
				{DType: "num", Num: float64(now + 10)},
				{DType: "str", Str: ASSIGNED_SURVEY_CATEGORY_NORMAL},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		action.Data[0].Str = "testSurveyKey2"
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.AssignedSurveys) != 3 {
			t.Errorf("unexpected number of surveys: %d", len(newState.PState.AssignedSurveys))
			return
		}

		// REMOVE_SURVEYS_BY_KEY
		action = studyTypes.Expression{
			Name: "REMOVE_SURVEYS_BY_KEY",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testSurveyKey1"},
			},
		}
		newState, err = ActionEval(action, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if len(newState.PState.AssignedSurveys) != 1 {
			t.Errorf("unexpected number of surveys: %d", len(newState.PState.AssignedSurveys))
		}
	})

	t.Run("ADD_MESSAGE", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "ADD_MESSAGE",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testMessage"},
				{DType: "num", Num: float64(time.Now().Unix() - 10)},
			},
		}
		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.Messages) != 1 {
			t.Errorf("unexpected number of messages: %d", len(newState.PState.Messages))
			return
		}
		if newState.PState.Messages[0].Type != "testMessage" {
			t.Errorf("unexpected message: %s", newState.PState.Messages[0].Type)
		}

		action2 := studyTypes.Expression{
			Name: "ADD_MESSAGE",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testMessage2"},
				{DType: "exp", Exp: &studyTypes.Expression{
					Name: "timestampWithOffset",
					Data: []studyTypes.ExpressionArg{
						{DType: "num", Num: 0},
					},
				}},
			},
		}
		newState, err = ActionEval(action2, newState, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.Messages) != 2 {
			t.Errorf("unexpected number of messages: %d", len(newState.PState.Messages))
			return
		}
		if newState.PState.Messages[1].Type != "testMessage2" {
			t.Errorf("unexpected message: %s", newState.PState.Messages[0].Type)
		}
		actionData = newState
	})

	t.Run("REMOVE_ALL_MESSAGES", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "REMOVE_ALL_MESSAGES",
			Data: []studyTypes.ExpressionArg{},
		}

		if len(actionData.PState.Messages) != 2 {
			t.Errorf("unexpected number of messages: %d", len(actionData.PState.Messages))
		}

		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.Messages) != 0 {
			t.Errorf("unexpected number of messages: %d", len(newState.PState.Messages))
			return
		}
	})

	t.Run("REMOVE_MESSAGES_BY_TYPE", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "REMOVE_MESSAGES_BY_TYPE",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "testMessage1"},
			},
		}
		actionData.PState = studyTypes.Participant{
			Messages: []studyTypes.ParticipantMessage{
				{Type: "testMessage1"},
				{Type: "testMessage1"},
				{Type: "testMessage2"},
				{Type: "testMessage3"},
				{Type: "testMessage4"},
			},
		}

		newState, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if len(newState.PState.Messages) != 3 {
			t.Errorf("unexpected number of messages: %d", len(newState.PState.Messages))
			return
		}
	})
}

func TestReportActions(t *testing.T) {
	// Override Now function for testing
	originalNow := Now
	defer func() { Now = originalNow }()
	Now = func() time.Time {
		return time.Unix(1609459200, 0) // Fixed time for testing: 2021-01-01 00:00:00 UTC
	}

	actionData := ActionData{
		PState: studyTypes.Participant{
			ParticipantID: "participant1234",
			StudyStatus:   studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE,
		},
		ReportsToCreate: map[string]studyTypes.Report{},
	}
	event := StudyEvent{
		Type: "SUBMIT",
		Response: studyTypes.SurveyResponse{
			Key: "test",
		},
	}
	var err error
	t.Run("INIT_REPORT", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "INIT_REPORT",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "key1"},
			},
		}

		actionData, err = ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(actionData.ReportsToCreate) < 1 {
			t.Error("should have one report initialized")
		}
	})

	t.Run("UPDATE_REPORT_DATA with no report there yet", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "UPDATE_REPORT_DATA",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "key2"},
				{DType: "str", Str: "d1"},
				{DType: "str", Str: "v1"},
			},
		}
		_, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(actionData.ReportsToCreate) < 2 {
			t.Error("should have two report initialized")
			return
		}
		if len(actionData.ReportsToCreate["key2"].Data) < 1 || actionData.ReportsToCreate["key2"].Data[0].Value != "v1" {
			t.Error("unexpected report value")
		}
	})

	t.Run("UPDATE_REPORT_DATA test for expression argument missing", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "UPDATE_REPORT_DATA",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "key2"},
				{DType: "str", Str: "d1"},
				{DType: "exp"},
			},
		}
		_, err := ActionEval(action, actionData, event)
		if err == nil {
			t.Error("should return an error about missing expression argument")
			return
		}
	})

	t.Run("UPDATE_REPORT_DATA update existing report existing attribute", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "UPDATE_REPORT_DATA",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "key2"},
				{DType: "str", Str: "d1"},
				{DType: "str", Str: "v2"},
			},
		}
		_, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(actionData.ReportsToCreate) < 2 {
			t.Error("should have two report initialized")
			return
		}
		if len(actionData.ReportsToCreate["key2"].Data) < 1 || actionData.ReportsToCreate["key2"].Data[0].Value != "v2" {
			t.Error("unexpected report value")
		}
	})

	t.Run("REMOVE_REPORT_DATA", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "REMOVE_REPORT_DATA",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "key2"},
				{DType: "str", Str: "d1"},
			},
		}
		_, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(actionData.ReportsToCreate) < 2 {
			t.Error("should have two report initialized")
			return
		}
		if len(actionData.ReportsToCreate["key2"].Data) > 0 {
			t.Error("unexpected report value")
		}
	})

	t.Run("CANCEL_REPORT", func(t *testing.T) {
		action := studyTypes.Expression{
			Name: "CANCEL_REPORT",
			Data: []studyTypes.ExpressionArg{
				{DType: "str", Str: "key2"},
			},
		}
		_, err := ActionEval(action, actionData, event)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if len(actionData.ReportsToCreate) > 1 {
			t.Error("should have only one report initialized")
		}
	})
}
