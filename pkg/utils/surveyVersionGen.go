package utils

import (
	"fmt"
	"time"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

func GenerateSurveyVersionID(oldVersions []*studyTypes.Survey) string {
	t := time.Now()

	date := t.Format("06-01")

	counter := 1
	newID := fmt.Sprintf("%s-%d", date, counter)
	for {
		idAlreadyPresent := false
		for _, v := range oldVersions {
			if v.VersionID == newID {
				idAlreadyPresent = true
				break
			}
		}
		if !idAlreadyPresent {
			break
		} else {
			counter += 1
			newID = fmt.Sprintf("%s-%d", date, counter)
		}
	}

	return newID
}
