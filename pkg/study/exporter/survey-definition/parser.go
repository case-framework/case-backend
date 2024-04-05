package surveydefinition

import (
	"log/slog"

	studyTypes "github.com/case-framework/case-backend/pkg/study/study"

	"github.com/case-framework/case-backend/pkg/utils"
)

func SurveyDefToVersionPreview(
	original *studyTypes.Survey,
	options *ExtractOptions,
) SurveyVersionPreview {
	sp := SurveyVersionPreview{
		VersionID:   original.VersionID,
		Published:   original.Published,
		Unpublished: original.Unpublished,
		Questions:   []SurveyQuestion{},
	}

	sp.Questions = extractQuestions(&original.SurveyDefinition, options)
	return sp
}

func extractQuestions(
	root *studyTypes.SurveyItem,
	options *ExtractOptions,
) []SurveyQuestion {
	questions := []SurveyQuestion{}
	if root == nil {
		return questions
	}

	var includeItemNames []string
	var excludeItemNames []string
	useLabels := false
	var prefLang string
	if options != nil {
		includeItemNames = options.IncludeItems
		excludeItemNames = options.ExcludeItems
		useLabels = options.UseLabelLang != ""
		prefLang = options.UseLabelLang
	}

	for _, item := range root.Items {
		if item.Type == studyTypes.SURVEY_ITEM_TYPE_PAGE_BREAK || item.Type == studyTypes.SURVEY_ITEM_TYPE_END {
			continue
		}
		if item.ConfidentialMode != "" {
			continue
		}
		if isItemGroup(&item) {
			questions = append(questions, extractQuestions(&item, options)...)
			continue
		}

		if len(includeItemNames) > 0 {
			if !utils.ContainsString(includeItemNames, item.Key) {
				continue
			}
		} else if utils.ContainsString(excludeItemNames, item.Key) {
			continue
		}

		rg := getResponseGroupComponent(&item)
		if rg == nil {
			continue
		}

		responses, qType := extractResponses(rg, prefLang)

		title := ""
		if useLabels {
			titleComp := getTitleComponent(&item)

			if titleComp != nil {
				var err error
				title, err = getPreviewText(titleComp, prefLang)
				if err != nil {
					slog.Debug("could not find question title", slog.String("item key", item.Key), slog.String("error", err.Error()))
				}
			}
		}

		question := SurveyQuestion{
			ID:           item.Key,
			Title:        title,
			QuestionType: qType,
			Responses:    responses,
		}
		questions = append(questions, question)
	}
	return questions
}
