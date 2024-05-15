package study

import (
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/study/studyengine"
	"github.com/case-framework/case-backend/pkg/study/types"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

/* func checkIfParticipantExists(instanceID string, studyKey string, participantID string, withStatus string) bool {
	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil || (withStatus != "" && pState.StudyStatus != withStatus) {
		return false
	}
	return err == nil
} */

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

func isSurveyAssignedAndActive(pState studyTypes.Participant, surveyKey string) bool {
	now := time.Now().Unix()

	for _, as := range pState.AssignedSurveys {
		if as.SurveyKey != surveyKey {
			continue
		}

		if as.ValidFrom > 0 && now < as.ValidFrom {
			continue
		}

		if as.ValidUntil > 0 && now > as.ValidUntil {
			continue
		}

		// --> survey is currently active
		return true
	}

	return false
}
