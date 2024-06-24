package surveydefinition

import (
	"errors"
	"log/slog"
	"strings"

	studytypes "github.com/case-framework/case-backend/pkg/study/types"
)

func mapToResponseDef(rItem *studytypes.ItemComponent, lang string) []ResponseDef {
	if rItem == nil {
		slog.Error("unexpected nil input")
		return []ResponseDef{}
	}

	key := rItem.Key
	responseDef := ResponseDef{
		ID: key,
	}

	var itemRole string
	roleSeparatorIndex := strings.Index(rItem.Role, ":")

	if roleSeparatorIndex == -1 {
		itemRole = rItem.Role
	} else if roleSeparatorIndex == 0 {
		responseDef.ResponseType = QUESTION_TYPE_UNKNOWN
		return []ResponseDef{responseDef}
	} else {
		itemRole = rItem.Role[0:roleSeparatorIndex]
	}

	switch itemRole {
	case "singleChoiceGroup":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				slog.Debug("label not found for component")
			}

			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			switch o.Role {
			case "option":
				option.OptionType = OPTION_TYPE_RADIO
			case "input":
				option.OptionType = OPTION_TYPE_TEXT_INPUT
			case "dateInput":
				option.OptionType = OPTION_TYPE_DATE_INPUT
			case "timeInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "cloze":
				option.OptionType = OPTION_TYPE_CLOZE
			}
			responseDef.Options = append(responseDef.Options, option)
			if option.OptionType == OPTION_TYPE_CLOZE {
				clozeOptions := extractClozeInputOptions(o, option.ID, lang)
				responseDef.Options = append(responseDef.Options, clozeOptions...)
			}
		}
		responseDef.ResponseType = QUESTION_TYPE_SINGLE_CHOICE
		return []ResponseDef{responseDef}
	case "multipleChoiceGroup":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				slog.Debug("label not found for component")
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			switch o.Role {
			case "option":
				option.OptionType = OPTION_TYPE_CHECKBOX
			case "input":
				option.OptionType = OPTION_TYPE_TEXT_INPUT
			case "dateInput":
				option.OptionType = OPTION_TYPE_DATE_INPUT
			case "timeInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "cloze":
				option.OptionType = OPTION_TYPE_CLOZE
			}
			responseDef.Options = append(responseDef.Options, option)
			if option.OptionType == OPTION_TYPE_CLOZE {
				clozeOptions := extractClozeInputOptions(o, option.ID, lang)
				responseDef.Options = append(responseDef.Options, clozeOptions...)
			}
		}
		responseDef.ResponseType = QUESTION_TYPE_MULTIPLE_CHOICE
		return []ResponseDef{responseDef}
	case "dropDownGroup":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				slog.Debug("label not found for component")
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			option.OptionType = OPTION_TYPE_DROPDOWN_OPTION
			responseDef.Options = append(responseDef.Options, option)
		}
		responseDef.ResponseType = QUESTION_TYPE_DROPDOWN
		return []ResponseDef{responseDef}
	case "input":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_TEXT_INPUT
		return []ResponseDef{responseDef}
	case "validatedRandomQuestion":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_TEXT_INPUT
		return []ResponseDef{responseDef}
	case "consent":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_CONSENT
		return []ResponseDef{responseDef}
	case "multilineTextInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_TEXT_INPUT
		return []ResponseDef{responseDef}
	case "numberInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMBER_INPUT
		return []ResponseDef{responseDef}
	case "dateInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_DATE_INPUT
		return []ResponseDef{responseDef}
	case "timeInput":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMBER_INPUT
		return []ResponseDef{responseDef}
	case "eq5d-health-indicator":
		responseDef.Label = ""
		responseDef.ResponseType = QUESTION_TYPE_EQ5D_SLIDER
		return []ResponseDef{responseDef}
	case "sliderNumeric":
		label, err := getPreviewText(rItem, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		responseDef.Label = label
		responseDef.ResponseType = QUESTION_TYPE_NUMERIC_SLIDER
		return []ResponseDef{responseDef}
	case "likert":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				slog.Debug("label not found for component")
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			option.OptionType = OPTION_TYPE_RADIO
			responseDef.Options = append(responseDef.Options, option)
		}
		responseDef.ResponseType = QUESTION_TYPE_LIKERT
		return []ResponseDef{responseDef}
	case "likertGroup":
		responses := []ResponseDef{}
		for _, likertComp := range rItem.Items {
			if likertComp.Role != "likert" {
				continue
			}
			subKey := likertComp.Key
			currentResponseDef := ResponseDef{
				ID:           subKey,
				ResponseType: QUESTION_TYPE_LIKERT_GROUP,
			}
			for _, o := range likertComp.Items {
				option := ResponseOption{
					ID: o.Key,
				}
				option.OptionType = OPTION_TYPE_RADIO
				currentResponseDef.Options = append(currentResponseDef.Options, option)
			}
			responses = append(responses, currentResponseDef)
		}
		return responses
	case "responsiveSingleChoiceArray":
		responses := []ResponseDef{}

		var options *studytypes.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "options" {
				options = &item
				break
			}
			continue
		}
		if options == nil {
			slog.Debug("options not found in component")
			return responses
		}

		for _, slot := range rItem.Items {
			if slot.Role != "row" {
				continue
			}
			subKey := slot.Key

			label, err := getPreviewText(&slot, lang)
			if err != nil {
				slog.Debug("label not found for component")
			}

			currentResponseDef := ResponseDef{
				ID:           subKey,
				ResponseType: QUESTION_TYPE_RESPONSIVE_SINGLE_CHOICE_ARRAY,
				Label:        label,
			}
			for _, o := range options.Items {
				label, err := getPreviewText(&o, lang)
				if err != nil {
					slog.Debug("label not found for component")
				}

				option := ResponseOption{
					ID:    o.Key,
					Label: label,
				}
				option.OptionType = OPTION_TYPE_RADIO
				currentResponseDef.Options = append(currentResponseDef.Options, option)
			}
			responses = append(responses, currentResponseDef)
		}
		return responses
	case "responsiveBipolarLikertScaleArray":
		responses := []ResponseDef{}

		var options *studytypes.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "options" {
				options = &item
				break
			}
			continue
		}
		if options == nil {
			slog.Debug("options not found in component")
			return responses
		}

		for _, slot := range rItem.Items {
			if slot.Role != "row" {
				continue
			}
			subKey := slot.Key

			var start *studytypes.ItemComponent
			var end *studytypes.ItemComponent
			for _, item := range slot.Items {
				if start != nil && end != nil {
					break
				}
				if item.Role == "start" {
					start = &item
					continue
				} else if item.Role == "end" {
					end = &item
					continue
				}
			}

			startLabel, err := getPreviewText(start, lang)
			if err != nil {
				slog.Debug("start label not found for component")
			}
			endLabel, err := getPreviewText(end, lang)
			if err != nil {
				slog.Debug("end label not found for component")
			}

			currentResponseDef := ResponseDef{
				ID:           subKey,
				ResponseType: QUESTION_TYPE_RESPONSIVE_BIPOLAR_LIKERT_ARRAY,
				Label:        startLabel + " vs. " + endLabel,
			}
			for _, o := range options.Items {
				option := ResponseOption{
					ID:    o.Key,
					Label: o.Key,
				}
				option.OptionType = OPTION_TYPE_RADIO
				currentResponseDef.Options = append(currentResponseDef.Options, option)
			}
			responses = append(responses, currentResponseDef)
		}
		return responses
	case "matrix":
		responses := []ResponseDef{}
		for _, row := range rItem.Items {
			rowKey := key + "." + row.Key
			if row.Role == "responseRow" {
				for _, col := range row.Items {
					cellKey := rowKey + "." + col.Key
					currentResponseDef := ResponseDef{
						ID: cellKey,
					}
					if col.Role == "dropDownGroup" {
						for _, o := range col.Items {
							dL, err := getPreviewText(&o, lang)
							if err != nil {
								slog.Debug("label not found for component")
							}
							option := ResponseOption{
								ID:    o.Key,
								Label: dL,
							}
							option.OptionType = OPTION_TYPE_DROPDOWN_OPTION
							currentResponseDef.Options = append(currentResponseDef.Options, option)
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_DROPDOWN
					} else if col.Role == "input" {
						label, err := getPreviewText(&col, lang)
						if err != nil {
							slog.Debug("label not found for component")
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_INPUT
						currentResponseDef.Label = label
					} else if col.Role == "check" {
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_CHECKBOX
					} else if col.Role == "numberInput" {
						label, err := getPreviewText(&col, lang)
						if err != nil {
							slog.Debug("label not found for component")
						}
						currentResponseDef.ResponseType = QUESTION_TYPE_MATRIX_NUMBER_INPUT
						currentResponseDef.Label = label
					} else {
						slog.Debug("matrix cell role ignored", slog.String("role", col.Role), slog.String("key", col.Key))
						continue
					}
					responses = append(responses, currentResponseDef)
				}
			} else if row.Role == "radioRow" {
				currentResponseDef := ResponseDef{
					ID:           rowKey,
					ResponseType: QUESTION_TYPE_MATRIX_RADIO_ROW,
				}
				for _, o := range row.Items {
					if o.Role == "label" {
						label, err := getPreviewText(&o, lang)
						if err != nil {
							slog.Debug("label not found for component")
						}
						currentResponseDef.Label = label
					} else {
						option := ResponseOption{
							ID: o.Key,
						}
						option.OptionType = OPTION_TYPE_RADIO
						currentResponseDef.Options = append(currentResponseDef.Options, option)
					}
				}
				responses = append(responses, currentResponseDef)
			}
		}
		return responses
	case "responsiveMatrix":
		responses := []ResponseDef{}

		var columns *studytypes.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "columns" {
				columns = &item
				break
			}
			continue
		}
		if columns == nil {
			slog.Debug("responsiveMatrix - columns not found in component")
			return responses
		}

		var rows *studytypes.ItemComponent
		for _, item := range rItem.Items {
			if item.Role == "rows" {
				rows = &item
				break
			}
			continue
		}
		if rows == nil {
			slog.Debug("responsiveMatrix - rows not found in component")
			return responses
		}

		for _, row := range rows.Items {
			if row.Role == "category" {
				// ignore category rows
				continue
			}
			rowLabel, err := getPreviewText(&row, lang)
			if err != nil {
				slog.Debug("row label not found for component")
			}

			for _, col := range columns.Items {
				slotKey := row.Key + "-" + col.Key

				colLabel, err := getPreviewText(&col, lang)
				if err != nil {
					slog.Debug("column label not found for component")
				}
				currentResponseDef := ResponseDef{
					ID:           slotKey,
					ResponseType: QUESTION_TYPE_RESPONSIVE_TABLE,
					Label:        rowLabel + " || " + colLabel,
				}
				responses = append(responses, currentResponseDef)
			}
		}
		return responses
	case "cloze":
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				slog.Debug("label not found for component")
			}
			option := ResponseOption{
				ID:    o.Key,
				Label: label,
			}
			switch o.Role {
			case "input":
				option.OptionType = OPTION_TYPE_TEXT_INPUT
			case "dateInput":
				option.OptionType = OPTION_TYPE_DATE_INPUT
			case "timeInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "numberInput":
				option.OptionType = OPTION_TYPE_NUMBER_INPUT
			case "dropDownGroup":
				option.OptionType = OPTION_TYPE_DROPDOWN
			}
			if option.OptionType != "" {
				responseDef.Options = append(responseDef.Options, option)
			}
		}
		responseDef.ResponseType = QUESTION_TYPE_CLOZE
		return []ResponseDef{responseDef}
	case "contact":
		responses := []ResponseDef{}
		for _, o := range rItem.Items {
			label, err := getPreviewText(&o, lang)
			if err != nil {
				slog.Debug("label not found for component")
			}

			switch o.Role {
			case "fullName":
				responses = append(responses, ResponseDef{
					ID:           rItem.Key + "." + o.Key,
					ResponseType: QUESTION_TYPE_TEXT_INPUT,
					Label:        label,
				})
			case "email":
				responses = append(responses, ResponseDef{
					ID:           rItem.Key + "." + o.Key,
					ResponseType: QUESTION_TYPE_TEXT_INPUT,
					Label:        label,
				})
			case "phone":
				responses = append(responses, ResponseDef{
					ID:           rItem.Key + "." + o.Key,
					ResponseType: QUESTION_TYPE_TEXT_INPUT,
					Label:        label,
				})
			case "address":
				responses = append(responses, ResponseDef{
					ID:           rItem.Key + "." + "street",
					ResponseType: QUESTION_TYPE_TEXT_INPUT,
					Label:        "Street",
				})
				responses = append(responses, ResponseDef{
					ID:           rItem.Key + "." + "street2",
					ResponseType: QUESTION_TYPE_TEXT_INPUT,
					Label:        "Street 2",
				})
				responses = append(responses, ResponseDef{
					ID:           rItem.Key + "." + "city",
					ResponseType: QUESTION_TYPE_TEXT_INPUT,
					Label:        "City",
				})
				responses = append(responses, ResponseDef{
					ID:           rItem.Key + "." + "postalCode",
					ResponseType: QUESTION_TYPE_TEXT_INPUT,
					Label:        "Postal Code",
				})
			}
		}
		return responses
	default:
		if roleSeparatorIndex > 0 {
			responseDef.ResponseType = QUESTION_TYPE_UNKNOWN
			return []ResponseDef{responseDef}
		}
		slog.Debug("mapToResponseDef: component with role is ignored", slog.String("role", rItem.Role), slog.String("key", key))

		return []ResponseDef{}
	}
}

func getPreviewText(item *studytypes.ItemComponent, lang string) (string, error) {
	if lang == "ignored" || lang == "" {
		return "", nil
	}
	if item == nil {
		return "", errors.New("getPreviewText: item nil")
	}
	if len(item.Items) > 0 {
		translation := ""
		for _, item := range item.Items {
			part, _ := getTranslation(&item.Content, lang)
			translation += part
		}
		if translation == "" {
			return "", errors.New("translation missing")
		}
		return translation, nil
	} else {
		return getTranslation(&item.Content, lang)
	}
}

func getTranslation(content *[]studytypes.LocalisedObject, lang string) (string, error) {
	if len(*content) < 1 {
		return "", errors.New("translations missing")
	}

	for _, translation := range *content {
		if translation.Code == lang {
			mergedText := ""
			for _, p := range translation.Parts {
				curVal := p.Str
				if p.DType == "exp" {
					curVal = "<exp>"
				} else if p.DType == "num" {
					curVal = "<num>"
				}
				mergedText += curVal
			}
			return mergedText, nil
		}
	}
	return "", errors.New("translation missing")
}

func extractClozeInputOptions(option studytypes.ItemComponent, clozeKey string, lang string) (clozeoptions []ResponseOption) {
	clozeInputs := []ResponseOption{}
	for _, o := range option.Items {
		label, err := getPreviewText(&o, lang)
		if err != nil {
			slog.Debug("label not found for component")
		}
		option := ResponseOption{
			ID:    clozeKey + "." + o.Key,
			Label: label,
		}
		switch o.Role {
		case "input":
			option.OptionType = OPTION_TYPE_EMBEDDED_CLOZE_TEXT_INPUT
		case "dateInput":
			option.OptionType = OPTION_TYPE_EMBEDDED_CLOZE_DATE_INPUT
		case "timeInput":
			option.OptionType = OPTION_TYPE_EMBEDDED_CLOZE_NUMBER_INPUT
		case "numberInput":
			option.OptionType = OPTION_TYPE_EMBEDDED_CLOZE_NUMBER_INPUT
		case "dropDownGroup":
			option.OptionType = OPTION_TYPE_EMBEDDED_CLOZE_DROPDOWN
		}
		if option.OptionType != "" {
			clozeInputs = append(clozeInputs, option)
		}
	}
	return clozeInputs
}
