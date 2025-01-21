package surveyresponses

import (
	"log/slog"

	sd "github.com/case-framework/case-backend/pkg/study/exporter/survey-definition"
	studytypes "github.com/case-framework/case-backend/pkg/study/types"
)

func isEmbeddedCloze(optionType string) bool {
	return optionType == sd.OPTION_TYPE_EMBEDDED_CLOZE_DATE_INPUT || optionType == sd.OPTION_TYPE_EMBEDDED_CLOZE_DROPDOWN ||
		optionType == sd.OPTION_TYPE_EMBEDDED_CLOZE_NUMBER_INPUT || optionType == sd.OPTION_TYPE_EMBEDDED_CLOZE_TEXT_INPUT
}

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
				responseCols[valueKey] = selection.Value
			}
		}
	}
	return responseCols
}

func parseSingleChoiceGroupList(questionKey string, responseSlotDefs []sd.ResponseDef, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
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
			responseCols[valueKey] = selection.Value
		}
	}
	return responseCols
}

func parseSimpleMultipleChoiceGroup(questionKey string, responseSlotDef sd.ResponseDef, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Find responses
	rGroup := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+responseSlotDef.ID)
	if rGroup != nil {
		if len(rGroup.Items) > 0 {
			for _, option := range responseSlotDef.Options {
				responseCols[questionKey+questionOptionSep+option.ID] = sd.FALSE_VALUE
				if isEmbeddedCloze(option.OptionType) {
					responseCols[questionKey+questionOptionSep+option.ID] = ""
				}
			}

			for _, item := range rGroup.Items {
				responseCols[questionKey+questionOptionSep+item.Key] = sd.TRUE_VALUE
				valueKey := questionKey + questionOptionSep + item.Key

				// Check if selected option is a cloze option
				cloze := false
				for _, option := range responseSlotDef.Options {
					if option.ID == item.Key && option.OptionType == sd.OPTION_TYPE_CLOZE {
						cloze = true
					}
				}

				// Handle cloze option specifically if we found it
				if cloze {
					for _, item := range item.Items {
						key := valueKey + "." + item.Key

						// Check if cloze or similar data structure
						if item.Value == "" && len(item.Items) == 1 {
							responseCols[key] = item.Items[0].Key
						} else {
							responseCols[key] = item.Value
						}
					}
				} else {
					valueKey += questionOptionSep + sd.OPEN_FIELD_COL_SUFFIX
					responseCols[valueKey] = item.Value
				}
			}
		}
	}

	return responseCols
}

func parseMultipleChoiceGroupList(questionKey string, responseSlotDefs []sd.ResponseDef, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Prepare columns:
	for _, rSlot := range responseSlotDefs {
		// Find responses
		rGroup := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+rSlot.ID)
		slotKeyPrefix := questionKey + questionOptionSep + rSlot.ID + "."
		if rGroup != nil {
			if len(rGroup.Items) > 0 {
				for _, option := range rSlot.Options {
					responseCols[slotKeyPrefix+option.ID] = sd.FALSE_VALUE
					if isEmbeddedCloze(option.OptionType) {
						responseCols[questionKey+questionOptionSep+option.ID] = ""
					}
				}

				for _, item := range rGroup.Items {
					responseCols[slotKeyPrefix+item.Key] = sd.TRUE_VALUE
					valueKey := slotKeyPrefix + item.Key

					// Check if selected option is a cloze option
					cloze := false
					for _, option := range rSlot.Options {
						if option.ID == item.Key && option.OptionType == sd.OPTION_TYPE_CLOZE {
							cloze = true
						}
					}

					// Handle cloze option specifically if we found it
					if cloze {
						for _, item := range item.Items {
							key := valueKey + "." + item.Key

							// Check if cloze or similar data structure
							if item.Value == "" && len(item.Items) == 1 {
								responseCols[key] = item.Items[0].Key
							} else {
								responseCols[key] = item.Value
							}
						}
					} else {
						valueKey += questionOptionSep + sd.OPEN_FIELD_COL_SUFFIX
						responseCols[valueKey] = item.Value
					}
				}
			}
		}
	}

	return responseCols
}

func parseSimpleCloze(questionKey string, responseSlotDef sd.ResponseDef, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Find responses
	rGroup := retrieveResponseItem(response, sd.RESPONSE_ROOT_KEY+"."+responseSlotDef.ID)
	if rGroup != nil {
		for _, item := range rGroup.Items {
			valueKey := questionKey + questionOptionSep + item.Key

			dropdown := false

			// Check if dropdown
			for _, option := range responseSlotDef.Options {
				if option.ID == item.Key && option.OptionType == sd.OPTION_TYPE_DROPDOWN {
					dropdown = true
					break
				}
			}

			if dropdown {
				if len(item.Items) != 1 {
					slog.Debug("multiple responses for dropdown in cloze", slog.String("questionKey", questionKey), slog.String("itemKey", item.Key))
				} else {
					responseCols[valueKey] = item.Items[0].Key
				}
			} else {
				responseCols[valueKey] = item.Value
			}
		}
	}
	return responseCols
}

func parseClozeList(questionKey string, responseSlotDefs []sd.ResponseDef, response *studytypes.SurveyItemResponse, questionOptionSep string) map[string]interface{} {
	responseCols := map[string]interface{}{}

	// Find responses:
	for _, rSlot := range responseSlotDefs {
		rGroup := retrieveResponseItemByShortKey(response, rSlot.ID)
		if rGroup == nil {
			continue
		}
		for _, item := range rGroup.Items {
			valueKey := questionKey + questionOptionSep + rSlot.ID + "." + item.Key

			dropdown := false

			// Check if dropdown
			for _, option := range rSlot.Options {
				if option.ID == item.Key && option.OptionType == sd.OPTION_TYPE_DROPDOWN {
					dropdown = true
					break
				}
			}

			if dropdown {
				if len(item.Items) != 1 {
					slog.Debug("multiple responses for dropdown in cloze", slog.String("questionKey", questionKey), slog.String("responseSlotID", rSlot.ID), slog.String("itemKey", item.Key))
				} else {
					responseCols[valueKey] = item.Items[0].Key
				}
			} else {
				responseCols[valueKey] = item.Value
			}
		}
	}
	return responseCols
}
