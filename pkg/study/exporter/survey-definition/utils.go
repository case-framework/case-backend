package surveydefinition

import (
	"strings"

	studyTypes "github.com/case-framework/case-backend/pkg/study/study"
)

func isItemGroup(item *studyTypes.SurveyItem) bool {
	return item != nil && len(item.Items) > 0
}

func getResponseGroupComponent(question *studyTypes.SurveyItem) *studyTypes.ItemComponent {
	if question.Components == nil {
		return nil
	}
	for _, c := range question.Components.Items {
		if c.Role == SURVEY_ITEM_COMPONENT_ROLE_RESPONSE_GROUP {
			return &c
		}
	}
	return nil
}

func getTitleComponent(question *studyTypes.SurveyItem) *studyTypes.ItemComponent {
	if question.Components == nil {
		return nil
	}
	for _, c := range question.Components.Items {
		if c.Role == SURVEY_ITEM_COMPONENT_ROLE_TITLE {
			return &c
		}
	}
	return nil
}

func extractResponses(rg *studyTypes.ItemComponent, lang string) ([]ResponseDef, string) {
	if rg == nil {
		return []ResponseDef{}, QUESTION_TYPE_EMPTY
	}

	responses := []ResponseDef{}
	for _, item := range rg.Items {
		r := mapToResponseDef(&item, lang)
		responses = append(responses, r...)

	}

	qType := getQuestionType(responses)
	return responses, qType

}

func getQuestionType(responses []ResponseDef) string {
	var qType string
	if len(responses) < 1 {
		qType = QUESTION_TYPE_EMPTY
	} else if len(responses) == 1 {
		qType = responses[0].ResponseType
	} else {
		// mixed or map to something specific (e.g., if all the same...)
		qType = responses[0].ResponseType

		// Check for matrix questions:
		if strings.Contains(qType, QUESTION_TYPE_MATRIX) {
			return QUESTION_TYPE_MATRIX
		}

		// Check for other questions, that contain same subtype
		for _, r := range responses {
			if qType != r.ResponseType {
				return QUESTION_TYPE_UNKNOWN
			}
		}
	}

	return qType
}
