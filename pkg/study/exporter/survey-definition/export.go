package surveydefinition

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

type SurveyInfoExporter struct {
	surveyKey   string
	surveyInfos []SurveyVersionPreview
}

func NewSurveyInfoExporter(surveyInfos []SurveyVersionPreview, surveyKey string, shortKeys bool) SurveyInfoExporter {
	if shortKeys {
		surveyInfos = removeRootKey(surveyInfos, surveyKey)
	}
	return SurveyInfoExporter{
		surveyInfos: surveyInfos,
		surveyKey:   surveyKey,
	}
}

func (se SurveyInfoExporter) GetSurveyInfos() []SurveyVersionPreview {
	return se.surveyInfos
}

func (se SurveyInfoExporter) GetSurveyInfoCSV(writer io.Writer) error {
	header := []string{
		"surveyKey", "versionID", "questionKey", "title",
		"responseKey", "type", "optionKey", "optionType", "optionLabel",
	}

	// Init writer
	w := csv.NewWriter(writer)

	// Write header
	err := w.Write(header)
	if err != nil {
		return err
	}

	for i, currentVersion := range se.surveyInfos {
		version := currentVersion.VersionID
		if version == "" {
			version = fmt.Sprintf("%d", i)
		}

		for _, question := range currentVersion.Questions {
			questionCols := []string{
				se.surveyKey,
				version,
				question.ID,
				question.Title,
			}
			for _, slot := range question.Responses {
				slotCols := []string{
					slot.ID,
					slot.ResponseType,
				}

				if len(slot.Options) > 0 {
					for _, option := range slot.Options {
						line := []string{}
						line = append(line, questionCols...)
						line = append(line, slotCols...)
						line = append(line, []string{
							option.ID,
							option.OptionType,
							option.Label,
						}...)

						err := w.Write(line)
						if err != nil {
							return err
						}
					}
				} else {
					line := []string{}
					line = append(line, questionCols...)
					line = append(line, slotCols...)
					line = append(line, []string{
						"",
						"",
						"",
					}...)
					err := w.Write(line)
					if err != nil {
						return err
					}
				}

			}
		}
	}

	w.Flush()
	return nil
}

func removeRootKey(surveyInfos []SurveyVersionPreview, surveyKey string) []SurveyVersionPreview {
	rootKey := surveyKey
	if rootKey == "" {
		return surveyInfos
	}

	for versionInd, sv := range surveyInfos {
		for qInd, question := range sv.Questions {
			surveyInfos[versionInd].Questions[qInd].ID = strings.TrimPrefix(question.ID, rootKey+".")
		}
	}
	return surveyInfos
}
