package study

import (
	studydb "github.com/case-framework/case-backend/pkg/db/study"
)

var (
	studyDBService *studydb.StudyDBService
)

func Init(studyDB *studydb.StudyDBService) {
	studyDBService = studyDB
}
