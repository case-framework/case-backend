package surveyresponses

import (
	"log/slog"

	sd "github.com/case-framework/case-backend/pkg/exporter/survey-definition"
	studytypes "github.com/case-framework/case-backend/pkg/types/study"
)

func parseSimpleSingleChoiceGroup(questionKey string, responseSlotDef sd.ResponseDef, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Find responses
	rGroup := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+responseSlotDef.ID)
	if rGroup != nil {
		if len(rGroup.Items) != 1 {
			slog.Debug("unexpected response group for question", slog.String("questionKey", questionKey), slog.Any("responseGroup", rGroup))
		} else {
			selection := rGroup.Items[0]
			responseCols[questionKey] = selection.Key
			valueKey := questionKey + questionOptionSep + selection.Key

			// Check if selected option is a cloze option
			cloze := false
			for _, option := range responseSlotDef.Options {
				if option.ID == selection.Key && option.OptionType == sd.OPTION_TYPE_CLOZE {
					cloze = true
					break
				}
			}

			// Handle cloze option specifically if we found it
			if cloze {
				for _, item := range selection.Items {
					key := valueKey + "." + item.Key

					// Check if cloze or similar data structure
					if item.Value == "" && len(item.Items) == 1 {
						responseCols[key] = item.Items[0].Key
					} else {
						responseCols[key] = item.Value
					}
				}
			} else {
				if _, hasKey := responseCols[valueKey]; hasKey {
					responseCols[valueKey] = selection.Value
				}
			}
		}
	}
	return responseCols
}

func handleSingleChoiceGroupList(questionKey string, responseSlotDefs []sd.ResponseDef, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Find responses:
	for _, rSlot := range responseSlotDefs {
		rGroup := retrieveResponseItemByShortKey(response, rSlot.ID)
		if rGroup == nil {
			continue
		} else if len(rGroup.Items) != 1 {
			slog.Debug("unexpected response group for question", slog.String("questionKey", questionKey), slog.Any("responseGroup", rGroup))
			continue
		}

		selection := rGroup.Items[0]
		responseCols[questionKey+questionOptionSep+rSlot.ID] = selection.Key
		valueKey := questionKey + questionOptionSep + rSlot.ID + "." + selection.Key

		// Check if selected option is a cloze option
		cloze := false
		for _, option := range rSlot.Options {
			if option.ID == selection.Key && option.OptionType == sd.OPTION_TYPE_CLOZE {
				cloze = true
				break
			}
		}

		// Handle cloze option specifically if we found it
		if cloze {
			for _, item := range selection.Items {
				key := valueKey + "." + item.Key

				// Check if cloze or similar data structure
				if item.Value == "" && len(item.Items) == 1 {
					responseCols[key] = item.Items[0].Key
				} else {
					responseCols[key] = item.Value
				}
			}
		} else {
			if _, hasKey := responseCols[valueKey]; hasKey {
				responseCols[valueKey] = selection.Value
			}
		}
	}
	return responseCols
}
