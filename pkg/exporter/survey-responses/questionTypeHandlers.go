package surveyresponses

import (
	sd "github.com/case-framework/case-backend/pkg/exporter/survey-definition"
	studytypes "github.com/case-framework/case-backend/pkg/types/study"
)

type QuestionTypeHandler interface {
	GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string
	ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{}
}

var questionTypeHandlers = map[string]QuestionTypeHandler{
	sd.QUESTION_TYPE_SINGLE_CHOICE:   &SingleChoiceHandler{},
	sd.QUESTION_TYPE_MULTIPLE_CHOICE: &MultipleChoiceHandler{},
	sd.QUESTION_TYPE_CONSENT:         &ConsentHandler{},
	sd.QUESTION_TYPE_DROPDOWN:        &SingleChoiceHandler{},
	sd.QUESTION_TYPE_LIKERT:          &SingleChoiceHandler{},
	sd.QUESTION_TYPE_LIKERT_GROUP:    &SingleChoiceHandler{},
	// TODO: add more handlers for other question types here
}

// SingleChoiceHandler implements the QuestionTypeHandler interface for single choice questions
type SingleChoiceHandler struct{}

func (h *SingleChoiceHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	cols := []string{}
	questionKey := question.ID
	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]

		cols = append(cols, question.ID)
		for _, option := range rSlot.Options {
			if option.OptionType != sd.OPTION_TYPE_RADIO &&
				option.OptionType != sd.OPTION_TYPE_DROPDOWN_OPTION &&
				option.OptionType != sd.OPTION_TYPE_CLOZE {
				cols = append(cols, question.ID+questionOptionSep+option.ID)
			}
		}
	} else {
		for _, rSlot := range question.Responses {
			cols = append(cols, questionKey+questionOptionSep+rSlot.ID)
			for _, option := range rSlot.Options {
				if option.OptionType != sd.OPTION_TYPE_RADIO &&
					option.OptionType != sd.OPTION_TYPE_DROPDOWN_OPTION && option.OptionType != sd.OPTION_TYPE_CLOZE {
					cols = append(cols, questionKey+questionOptionSep+rSlot.ID+"."+option.ID)
				}
			}
		}
	}

	return cols
}

func (h *SingleChoiceHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	var responseCols map[string]interface{}

	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		responseCols = parseSimpleSingleChoiceGroup(question.ID, rSlot, response, questionOptionSep)
	} else {
		responseCols = parseSingleChoiceGroupList(question.ID, question.Responses, response, questionOptionSep)
	}
	return responseCols
}

// MultipleChoiceHandler implements the QuestionTypeHandler interface for multiple choice questions
type MultipleChoiceHandler struct{}

func (h *MultipleChoiceHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	cols := []string{}

	questionKey := question.ID
	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]

		for _, option := range rSlot.Options {
			colName := questionKey + questionOptionSep + option.ID
			cols = append(cols, colName)

			if option.OptionType != sd.OPTION_TYPE_CHECKBOX && option.OptionType != sd.OPTION_TYPE_CLOZE && !isEmbeddedCloze(option.OptionType) {
				colName := questionKey + questionOptionSep + option.ID + questionOptionSep + sd.OPEN_FIELD_COL_SUFFIX
				cols = append(cols, colName)
			}
		}
	} else {
		for _, rSlot := range question.Responses {
			slotKeyPrefix := questionKey + questionOptionSep + rSlot.ID + "."

			for _, option := range rSlot.Options {
				colName := slotKeyPrefix + option.ID
				cols = append(cols, colName)

				if option.OptionType != sd.OPTION_TYPE_CHECKBOX && option.OptionType != sd.OPTION_TYPE_CLOZE && !isEmbeddedCloze(option.OptionType) {
					colName := slotKeyPrefix + option.ID + questionOptionSep + sd.OPEN_FIELD_COL_SUFFIX
					cols = append(cols, colName)
				}
			}
		}
	}

	return cols
}

func (h *MultipleChoiceHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	var responseCols map[string]interface{}

	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		responseCols = parseSimpleMultipleChoiceGroup(question.ID, rSlot, response, questionOptionSep)

	} else {
		responseCols = parseMultipleChoiceGroupList(question.ID, question.Responses, response, questionOptionSep)
	}

	return responseCols
}

// ConsentHandler implements the QuestionTypeHandler interface for consent questions
type ConsentHandler struct{}

func (h *ConsentHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	responseCols := []string{}
	if len(question.Responses) == 1 {
		responseCols = append(responseCols, question.ID)
	} else {
		for _, rSlot := range question.Responses {
			responseCols = append(responseCols, question.ID+questionOptionSep+rSlot.ID)
		}
	}
	return responseCols
}

func (h *ConsentHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	questionKey := question.ID
	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		rValue := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+rSlot.ID)
		if rValue != nil {
			responseCols[questionKey] = sd.TRUE_VALUE
		} else {
			responseCols[questionKey] = sd.FALSE_VALUE
		}

	} else {
		for _, rSlot := range question.Responses {
			// Prepare columns:
			slotKey := questionKey + questionOptionSep + rSlot.ID

			// Find responses
			rValue := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+rSlot.ID)
			if rValue != nil {
				responseCols[slotKey] = sd.TRUE_VALUE
			} else {
				responseCols[slotKey] = sd.FALSE_VALUE
			}
		}
	}
	return responseCols
}
