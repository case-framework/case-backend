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
	sd.QUESTION_TYPE_SINGLE_CHOICE: &SingleChoiceHandler{},
	// TODO: add more handlers for other question types here
}

// SingleChoiceHandler implements the QuestionTypeHandler interface for single choice questions
type SingleChoiceHandler struct{}

func (h *SingleChoiceHandler) GetResponseColumnNames(question sd.SurveyQuestion, questionOptionSep string) []string {
	cols := []string{}
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
			questionKey := question.ID
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
		responseCols = handleSingleChoiceGroupList(question.ID, question.Responses, response, questionOptionSep)
	}
	return responseCols
}
