package study

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	"go.mongodb.org/mongo-driver/bson"
)

type AssignedSurveyWithContext struct {
	Survey  *studyTypes.Survey         `json:"survey"`
	Context *SurveyContext             `json:"context,omitempty" `
	Prefill *studyTypes.SurveyResponse `json:"prefill,omitempty"`
}

type SurveyContext struct {
	Mode              string                      `json:"mode,omitempty"`
	PreviousResponses []studyTypes.SurveyResponse `json:"previousResponses,omitempty"`
	ParticipantFlags  map[string]string           `json:"participantFlags,omitempty"`
}

type SurveyInfo struct {
	SurveyKey       string                       `json:"surveyKey"`
	StudyKey        string                       `json:"studyKey"`
	Name            []studyTypes.LocalisedObject `json:"name"`
	Description     []studyTypes.LocalisedObject `json:"description"`
	TypicalDuration []studyTypes.LocalisedObject `json:"typicalDuration"`
	VersionID       string                       `json:"versionID"`
}

type AssignedSurveysWithInfos struct {
	Surveys     []studyTypes.AssignedSurvey `json:"surveys"`
	SurveyInfos []*SurveyInfo               `json:"surveyInfos"`
}

type SubmissionEntry struct {
	ProfileID string `json:"profileID"`
	Timestamp int64  `json:"timestamp"`
	SurveyKey string `json:"surveyKey"`
	VersionID string `json:"versionID"`
}

type SubmissionHistory struct {
	Submissions []SubmissionEntry `json:"submissions"`
	SurveyInfos []*SurveyInfo     `json:"surveyInfos"`
}

func GetAssignedSurveys(instanceID string, studyKey string, profileIDs []string) (surveysWithInfos AssignedSurveysWithInfos, err error) {
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	surveysWithInfos = AssignedSurveysWithInfos{
		Surveys:     []studyTypes.AssignedSurvey{},
		SurveyInfos: []*SurveyInfo{},
	}

	for _, profileID := range profileIDs {
		participantID, _, err := ComputeParticipantIDs(study, profileID)
		if err != nil {
			slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
			continue
		}

		pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
		if err != nil {
			slog.Debug("Error getting participant state", slog.String("error", err.Error()))
			continue
		}

		if pState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE {
			slog.Error("Participant is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
			continue
		}

		for _, survey := range pState.AssignedSurveys {
			survey.ProfileID = profileID
			survey.StudyKey = studyKey
			surveysWithInfos.Surveys = append(surveysWithInfos.Surveys, survey)
		}
	}

	for _, survey := range surveysWithInfos.Surveys {
		// is not in the survey info list yet
		found := false
		for _, surveyInfo := range surveysWithInfos.SurveyInfos {
			if surveyInfo.SurveyKey == survey.SurveyKey {
				found = true
				break
			}
		}

		if !found {
			surveyDef, err := studyDBService.GetCurrentSurveyVersion(instanceID, studyKey, survey.SurveyKey)
			if err != nil {
				slog.Error("error getting survey definition", slog.String("error", err.Error()), slog.String("surveyKey", survey.SurveyKey))
				continue
			}
			surveyInfo := SurveyInfo{
				SurveyKey:       survey.SurveyKey,
				StudyKey:        studyKey,
				Name:            surveyDef.Props.Name,
				Description:     surveyDef.Props.Description,
				TypicalDuration: surveyDef.Props.TypicalDuration,
			}
			surveysWithInfos.SurveyInfos = append(surveysWithInfos.SurveyInfos, &surveyInfo)
		}
	}

	err = nil
	return
}

func GetAssignedSurveysForTempParticipant(instanceID string, studyKey string, participantID string) (surveysWithInfos AssignedSurveysWithInfos, err error) {
	_, err = getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		slog.Error("error getting participant state", slog.String("error", err.Error()))
		return
	}

	if pState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY {
		slog.Error("participant is not temporary", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		err = errors.New("participant is not temporary")
		return
	}

	surveysWithInfos = AssignedSurveysWithInfos{
		Surveys:     pState.AssignedSurveys,
		SurveyInfos: []*SurveyInfo{},
	}

	for _, survey := range pState.AssignedSurveys {
		// is not in the survey info list yet
		found := false
		for _, surveyInfo := range surveysWithInfos.SurveyInfos {
			if surveyInfo.SurveyKey == survey.SurveyKey {
				found = true
				break
			}
		}

		if !found {
			surveyDef, err := studyDBService.GetCurrentSurveyVersion(instanceID, studyKey, survey.SurveyKey)
			if err != nil {
				slog.Error("error getting survey definition", slog.String("error", err.Error()))
				continue
			}
			surveyInfo := SurveyInfo{
				SurveyKey:       survey.SurveyKey,
				StudyKey:        studyKey,
				Name:            surveyDef.Props.Name,
				Description:     surveyDef.Props.Description,
				TypicalDuration: surveyDef.Props.TypicalDuration,
			}
			surveysWithInfos.SurveyInfos = append(surveysWithInfos.SurveyInfos, &surveyInfo)
		}
	}
	return
}

func GetAssignedSurveyWithContext(instanceID string, studyKey string, surveyKey string, profileID string) (surveyWithContent AssignedSurveyWithContext, err error) {
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	surveyDef, err := studyDBService.GetCurrentSurveyVersion(instanceID, studyKey, surveyKey)
	if err != nil {
		slog.Error("error getting survey", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))
		return
	}

	participantID, _, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		// participant not found
		if surveyDef.AvailableFor == studyTypes.SURVEY_AVAILABLE_FOR_PUBLIC {
			pState.AssignedSurveys = []studyTypes.AssignedSurvey{
				{
					SurveyKey: surveyKey,
					Category:  "normal",
				},
			}
		} else {
			slog.Error("participant not found", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
			return
		}
	}

	if surveyDef.AvailableFor == studyTypes.SURVEY_AVAILABLE_FOR_PARTICIPANTS_IF_ASSIGNED {
		if !isSurveyAssignedAndActive(pState, surveyKey) {
			slog.Error("survey is not assigned or inactive", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("surveyKey", surveyKey))
			err = errors.New("survey is not assigned or inactive")
			return
		}
	}

	// Prepare context
	surveyContext, err := resolveContextRules(instanceID, studyKey, pState, surveyDef.ContextRules)
	if err != nil {
		slog.Error("error resolving context rules", slog.String("error", err.Error()))
		return
	}
	// Prepare prefill
	prefill, err := resolvePrefillRules(instanceID, studyKey, participantID, surveyDef.PrefillRules)
	if err != nil {
		slog.Error("error resolving prefill rules", slog.String("error", err.Error()))
		return
	}

	surveyDef.ContextRules = nil
	surveyDef.PrefillRules = nil

	surveyWithContent = AssignedSurveyWithContext{
		Survey:  surveyDef,
		Context: surveyContext,
		Prefill: prefill,
	}
	return
}

func GetSurveyWithContextForTempParticipant(instanceID string, studyKey string, surveyKey string, tempParticipantID string) (surveyWithContent AssignedSurveyWithContext, err error) {
	_, err = getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	surveyDef, err := studyDBService.GetCurrentSurveyVersion(instanceID, studyKey, surveyKey)
	if err != nil {
		slog.Error("error getting survey", slog.String("error", err.Error()))
		return
	}

	if tempParticipantID == "" {
		// check if survey is available for public
		if surveyDef.AvailableFor != studyTypes.SURVEY_AVAILABLE_FOR_PUBLIC {
			slog.Error("survey is not available for public", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))
			err = errors.New("survey is not available for public")
			return
		}
	} else {
		// check if survey is available for temporary participant
		requiredAllowedFor := []string{studyTypes.SURVEY_AVAILABLE_FOR_TEMPORARY_PARTICIPANTS, studyTypes.SURVEY_AVAILABLE_FOR_PUBLIC}
		allowed := false
		for _, allowedFor := range requiredAllowedFor {
			if surveyDef.AvailableFor == allowedFor {
				allowed = true
				break
			}
		}
		if !allowed {
			slog.Error("survey is not available for temporary participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))
			err = errors.New("survey is not available for temporary participant")
			return
		}
	}

	// Prepare context
	var surveyContext *SurveyContext
	if tempParticipantID != "" {
		pState, err2 := studyDBService.GetParticipantByID(instanceID, studyKey, tempParticipantID)
		if err2 != nil {
			err = err2
			slog.Error("error getting participant state", slog.String("error", err.Error()))
			return
		}

		if pState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_TEMPORARY {
			slog.Error("participant is not temporary", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", tempParticipantID))
			err = errors.New("participant is not temporary")
			return
		}

		surveyContext, err = resolveContextRules(instanceID, studyKey, pState, surveyDef.ContextRules)
		if err != nil {
			slog.Error("error resolving context rules", slog.String("error", err.Error()))
			return
		}
	}

	// Prepare prefill
	prefill, err := resolvePrefillRules(instanceID, studyKey, tempParticipantID, surveyDef.PrefillRules)
	if err != nil {
		slog.Error("error resolving prefill rules", slog.String("error", err.Error()))
		return
	}

	surveyDef.PrefillRules = nil
	surveyDef.ContextRules = nil

	surveyWithContent = AssignedSurveyWithContext{
		Survey:  surveyDef,
		Context: surveyContext,
		Prefill: prefill,
	}
	return
}

func resolveContextRules(instanceID string, studyKey string, pState studyTypes.Participant, contextRules *studyTypes.SurveyContextDef) (sCtx *SurveyContext, err error) {
	sCtx = &SurveyContext{
		ParticipantFlags: pState.Flags,
	}

	if contextRules == nil {
		return
	}

	// mode:
	if contextRules.Mode != nil {
		modeRule := contextRules.Mode
		switch modeRule.DType {
		case "exp":
			return sCtx, errors.New("expression arg type not supported yet")
		case "str":
			sCtx.Mode = modeRule.Str
		default:
			sCtx.Mode = modeRule.Str
		}
	}

	// previous responses:
	prevRespRules := contextRules.PreviousResponses
	for _, rule := range prevRespRules {
		slog.Error("previous responses not implemented yet", slog.String("rule", rule.Name), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
	}
	return sCtx, nil

}

func resolvePrefillRules(instanceID string, studyKey string, participantID string, rules []studyTypes.Expression) (prefills *studyTypes.SurveyResponse, err error) {
	if rules == nil || len(rules) < 1 {
		return nil, nil
	}

	prefills = &studyTypes.SurveyResponse{
		Responses: []studyTypes.SurveyItemResponse{},
	}

	lastSurveyCache := map[string]studyTypes.SurveyResponse{}
	for _, rule := range rules {
		switch rule.Name {
		case "PREFILL_SLOT_WITH_VALUE":
			if len(rule.Data) < 3 {
				slog.Error("not enough arguments in", slog.String("rule", rule.Name))
				continue
			}
			itemKey := rule.Data[0].Str
			slotKey := rule.Data[1].Str
			targetValue := rule.Data[2]

			prefillItem := studyTypes.SurveyItemResponse{
				Key: itemKey,
			}

			// Find item if already exits
			pItemIndex := -1
			for i, p := range prefills.Responses {
				if p.Key == itemKey {
					prefillItem = p
					pItemIndex = i
					break
				}
			}

			slotKeyParts := strings.Split(slotKey, ".")
			if len(slotKeyParts) < 1 {
				slog.Error("prefill rule has invalid slot key", slog.String("rule", rule.Name))
				return
			}

			respItem := prefillItem.Response
			if respItem == nil {
				respItem = &studyTypes.ResponseItem{Key: slotKeyParts[0], Items: []*studyTypes.ResponseItem{}}
			}

			var currentRespItem *studyTypes.ResponseItem
			for _, rKey := range slotKeyParts {
				if currentRespItem == nil {
					currentRespItem = respItem
					continue
				}

				found := false
				for _, item := range currentRespItem.Items {
					if item.Key == rKey {
						found = true
						currentRespItem = item
						break
					}
				}
				if !found {
					newItem := studyTypes.ResponseItem{Key: rKey, Items: []*studyTypes.ResponseItem{}}
					currentRespItem.Items = append(currentRespItem.Items, &newItem)
					currentRespItem = currentRespItem.Items[len(currentRespItem.Items)-1]
				}
			}

			if targetValue.DType == "num" {
				currentRespItem.Dtype = "number"
				currentRespItem.Value = fmt.Sprintf("%f", targetValue.Num)
			} else {
				currentRespItem.Value = targetValue.Str
			}
			prefillItem.Response = respItem

			if pItemIndex > -1 {
				prefills.Responses[pItemIndex] = prefillItem
			} else {
				prefills.Responses = append(prefills.Responses, prefillItem)
			}
		case "GET_LAST_SURVEY_ITEM":
			if len(rule.Data) < 2 {
				slog.Error("GET_LAST_SURVEY_ITEM must have at least two arguments")
				continue
			}
			if participantID == "" {
				slog.Error("participantID is required")
				continue
			}
			surveyKey := rule.Data[0].Str
			itemKey := rule.Data[1].Str
			since := int64(0)
			if len(rule.Data) == 3 {
				// look up responses that are not older than:
				since = time.Now().Unix() - int64(rule.Data[2].Num)
			}

			previousResp, ok := lastSurveyCache[surveyKey]
			if !ok {
				filter := bson.M{
					"participantID": participantID,
					"surveyKey":     surveyKey,
					"arrivedAt":     bson.M{"$gt": since},
				}
				resps, _, err := studyDBService.GetResponses(instanceID, studyKey, filter, bson.M{"arrivedAt": -1}, 1, 1)

				if err != nil || len(resps) < 1 {
					continue
				}
				lastSurveyCache[surveyKey] = resps[0]
				previousResp = resps[0]
			}

			for _, item := range previousResp.Responses {
				if item.Key == itemKey {
					prefills.Responses = append(prefills.Responses, item)
					break
				}
			}
		default:
			return prefills, fmt.Errorf("expression is not supported yet: %s", rule.Name)
		}
	}
	return prefills, nil
}

func GetLinkingCode(instanceID string, studyKey string, profileID string, key string) (value string, err error) {
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	participantID, _, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return
	}

	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		slog.Debug("Error getting participant state", slog.String("error", err.Error()))
		return
	}

	if pState.LinkingCodes == nil {
		slog.Debug("no linking codes found for participant", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		return
	}
	value, ok := pState.LinkingCodes[key]
	if !ok {
		slog.Debug("linking code for key not found", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID), slog.String("key", key))
		return
	}

	return value, nil
}

func GetSubmissionHistory(instanceID string, studyKey string, profileIDs []string, limit int64) (submissionHistory SubmissionHistory, err error) {
	study, err := getStudyIfActive(instanceID, studyKey)
	if err != nil {
		slog.Error("error getting study", slog.String("error", err.Error()))
		return
	}

	submissionHistory = SubmissionHistory{
		Submissions: []SubmissionEntry{},
		SurveyInfos: []*SurveyInfo{},
	}

	for _, profileID := range profileIDs {
		participantID, _, err := ComputeParticipantIDs(study, profileID)
		if err != nil {
			slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
			continue
		}

		pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
		if err != nil {
			slog.Debug("Error getting participant state", slog.String("error", err.Error()))
			continue
		}

		if pState.StudyStatus != studyTypes.PARTICIPANT_STUDY_STATUS_ACTIVE {
			slog.Error("Participant is not active", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
			continue
		}

		// fetch submissions
		responseInfos, _, err := studyDBService.GetResponseInfos(instanceID, studyKey, bson.M{"participantID": participantID}, 1, limit)
		if err != nil {
			slog.Error("Error getting response infos", slog.String("error", err.Error()))
			continue
		}

		for _, responseInfo := range responseInfos {
			submissionHistory.Submissions = append(submissionHistory.Submissions, SubmissionEntry{
				ProfileID: profileID,
				Timestamp: responseInfo.ArrivedAt,
				SurveyKey: responseInfo.Key,
				VersionID: responseInfo.VersionID,
			})
		}
	}

	for _, subEntry := range submissionHistory.Submissions {
		// is not in the survey info list yet
		found := false
		for _, surveyInfo := range submissionHistory.SurveyInfos {
			if surveyInfo.SurveyKey == subEntry.SurveyKey {
				found = true
				break
			}
		}

		if !found {
			surveyDef, err := studyDBService.GetSurveyVersion(instanceID, studyKey, subEntry.SurveyKey, subEntry.VersionID)
			if err != nil {
				slog.Error("error getting survey definition with specific version", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", subEntry.SurveyKey), slog.String("versionID", subEntry.VersionID))

				allVersions, err := studyDBService.GetSurveyVersions(instanceID, studyKey, subEntry.SurveyKey)
				if err != nil {
					slog.Error("error getting survey definition with all versions", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", subEntry.SurveyKey))
					continue
				}
				if len(allVersions) < 1 {
					slog.Error("no survey definition found", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", subEntry.SurveyKey))
					continue
				}
				surveyDef = allVersions[len(allVersions)-1]
			}
			surveyInfo := SurveyInfo{
				SurveyKey:       subEntry.SurveyKey,
				StudyKey:        studyKey,
				Name:            surveyDef.Props.Name,
				Description:     surveyDef.Props.Description,
				TypicalDuration: surveyDef.Props.TypicalDuration,
				VersionID:       subEntry.VersionID,
			}
			submissionHistory.SurveyInfos = append(submissionHistory.SurveyInfos, &surveyInfo)
		}
	}

	err = nil
	return
}
