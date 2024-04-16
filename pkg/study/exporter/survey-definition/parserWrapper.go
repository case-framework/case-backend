package surveydefinition

import studyTypes "github.com/case-framework/case-backend/pkg/study/types"

type StudyDB interface {
	GetSurveyVersions(instanceID string, studyKey string, surveyKey string) (surveys []*studyTypes.Survey, err error)
	GetSurveyVersion(instanceID string, studyKey string, surveyKey string, versionID string) (survey *studyTypes.Survey, err error)
}

func PrepareSurveyInfosFromDB(
	db StudyDB,
	instanceID string,
	studyKey string,
	surveyKey string,
	parserOptions *ExtractOptions,
) (surveyInfos []SurveyVersionPreview, err error) {
	surveyVersions, err := db.GetSurveyVersions(instanceID, studyKey, surveyKey)
	if err != nil {
		return nil, err
	}

	for _, survey := range surveyVersions {
		surveyDefinition, err := db.GetSurveyVersion(instanceID, studyKey, surveyKey, survey.VersionID)
		if err != nil {
			return nil, err
		}

		surveyInfo := SurveyDefToVersionPreview(surveyDefinition, parserOptions)
		surveyInfos = append(surveyInfos, surveyInfo)
	}

	return surveyInfos, nil
}
