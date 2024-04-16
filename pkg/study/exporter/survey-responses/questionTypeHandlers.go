package surveyresponses

import (
	"log/slog"

	sd "github.com/case-framework/case-backend/pkg/study/exporter/survey-definition"
	studytypes "github.com/case-framework/case-backend/pkg/study/types"
)

type QuestionTypeHandler interface {
	GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string
	ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{}
}

var questionTypeHandlers = map[string]QuestionTypeHandler{
	sd.QUESTION_TYPE_SINGLE_CHOICE:                   &SingleChoiceHandler{},
	sd.QUESTION_TYPE_MULTIPLE_CHOICE:                 &MultipleChoiceHandler{},
	sd.QUESTION_TYPE_CONSENT:                         &ConsentHandler{},
	sd.QUESTION_TYPE_DROPDOWN:                        &SingleChoiceHandler{},
	sd.QUESTION_TYPE_LIKERT:                          &SingleChoiceHandler{},
	sd.QUESTION_TYPE_LIKERT_GROUP:                    &SingleChoiceHandler{},
	sd.QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY:  &SingleChoiceHandler{},
	sd.QUESTION_TYPE_RESPONSIVE_BIPOLAR_LIKERT_ARRAY: &SingleChoiceHandler{},
	sd.QUESTION_TYPE_TEXT_INPUT:                      &InputValueHandler{},
	sd.QUESTION_TYPE_DATE_INPUT:                      &InputValueHandler{},
	sd.QUESTION_TYPE_NUMBER_INPUT:                    &InputValueHandler{},
	sd.QUESTION_TYPE_NUMERIC_SLIDER:                  &InputValueHandler{},
	sd.QUESTION_TYPE_EQ5D_SLIDER:                     &InputValueHandler{},
	sd.QUESTION_TYPE_RESPONSIVE_TABLE:                &ResponsiveTableHandler{},
	sd.QUESTION_TYPE_MATRIX:                          &MatrixHandler{},
	sd.QUESTION_TYPE_CLOZE:                           &ClozeHandler{},
	sd.QUESTION_TYPE_UNKNOWN:                         &UnknownTypeHandler{},
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

// InputValueHandler implements the QuestionTypeHandler interface for input value questions
type InputValueHandler struct{}

func (h *InputValueHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	colNames := []string{}
	if len(question.Responses) == 1 {
		colNames = append(colNames, question.ID)
	} else {
		for _, rSlot := range question.Responses {
			slotKey := question.ID + questionOptionSep + rSlot.ID
			colNames = append(colNames, slotKey)
		}
	}
	return colNames
}

func (h *InputValueHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}
	questionKey := question.ID
	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		rValue := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+rSlot.ID)
		if rValue != nil {
			responseCols[questionKey] = rValue.Value
		}

	} else {
		for _, rSlot := range question.Responses {
			// Prepare columns:
			slotKey := questionKey + questionOptionSep + rSlot.ID
			responseCols[slotKey] = ""

			// Find responses
			rValue := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+rSlot.ID)
			if rValue != nil {
				responseCols[slotKey] = rValue.Value
			}
		}
	}
	return responseCols
}

// ResponsiveTableHandler implements the QuestionTypeHandler interface for responsive table questions
type ResponsiveTableHandler struct{}

func (h *ResponsiveTableHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	colNames := []string{}

	for _, rSlot := range question.Responses {
		slotKey := question.ID + questionOptionSep + rSlot.ID
		colNames = append(colNames, slotKey)
	}

	return colNames
}

func (h *ResponsiveTableHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	for _, rSlot := range question.Responses {
		slotKey := question.ID + questionOptionSep + rSlot.ID

		rItem := retrieveResponseItemByShortKey(response, rSlot.ID)

		if rItem != nil {
			responseCols[slotKey] = rItem.Value
		}
	}

	return responseCols
}

// MatrixHandler implements the QuestionTypeHandler interface for matrix questions
type MatrixHandler struct{}

func (h *MatrixHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	colNames := []string{}

	for _, rSlot := range question.Responses {
		slotKey := question.ID + questionOptionSep + rSlot.ID
		colNames = append(colNames, slotKey)
	}

	return colNames
}

func (h *MatrixHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	for _, rSlot := range question.Responses {
		slotKey := question.ID + questionOptionSep + rSlot.ID

		rGroup := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+rSlot.ID)

		if rSlot.ResponseType == sd.QUESTION_TYPE_MATRIX_RADIO_ROW {
			if rGroup != nil {
				if len(rGroup.Items) != 1 {
					slog.Debug("unexpected response group for question", slog.String("questionKey", question.ID))
				} else {
					selection := rGroup.Items[0]
					responseCols[slotKey] = selection.Key
				}
			}
		} else {
			if rGroup != nil {
				if len(rGroup.Items) != 1 {
					slog.Debug("unexpected response group for question", slog.String("questionKey", question.ID))
				} else {
					selection := rGroup.Items[0]
					value := selection.Key
					if selection.Value != "" {
						value = selection.Value
					}
					responseCols[slotKey] = value
				}
			}
		}
	}

	return responseCols
}

// ClozeHandler implements the QuestionTypeHandler interface for cloze questions
type ClozeHandler struct{}

func (h *ClozeHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	colNames := []string{}

	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		for _, option := range rSlot.Options {
			if option.OptionType == sd.OPTION_TYPE_DATE_INPUT || option.OptionType == sd.OPTION_TYPE_NUMBER_INPUT || option.OptionType == sd.OPTION_TYPE_TEXT_INPUT || option.OptionType == sd.OPTION_TYPE_DROPDOWN {
				slotKey := question.ID + questionOptionSep + option.ID
				colNames = append(colNames, slotKey)
			}
		}

	} else {
		for _, rSlot := range question.Responses {
			for _, option := range rSlot.Options {
				if option.OptionType == sd.OPTION_TYPE_DATE_INPUT || option.OptionType == sd.OPTION_TYPE_NUMBER_INPUT || option.OptionType == sd.OPTION_TYPE_TEXT_INPUT || option.OptionType == sd.OPTION_TYPE_DROPDOWN {
					slotKey := question.ID + questionOptionSep + rSlot.ID + "." + option.ID
					colNames = append(colNames, slotKey)
				}
			}
		}
	}

	return colNames
}

func (h *ClozeHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	var responseCols map[string]interface{}

	if len(question.Responses) == 1 {
		rSlot := question.Responses[0]
		responseCols = parseSimpleCloze(question.ID, rSlot, response, questionOptionSep)

	} else {
		responseCols = parseClozeList(question.ID, question.Responses, response, questionOptionSep)
	}

	return responseCols
}

// UnknownTypeHandler implements the QuestionTypeHandler interface for unknown question types
type UnknownTypeHandler struct{}

func (h *UnknownTypeHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	colNames := []string{}

	for _, rSlot := range question.Responses {
		slotKey := question.ID + questionOptionSep + rSlot.ID
		colNames = append(colNames, slotKey)
	}

	return colNames
}

func (h *UnknownTypeHandler) ParseResponse(question sd.SurveyQuestion, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	for _, rSlot := range question.Responses {
		slotKey := question.ID + questionOptionSep + rSlot.ID

		rGroup := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+rSlot.ID)
		if rGroup != nil {
			responseCols[slotKey] = rGroup
		}
	}

	return responseCols
}
