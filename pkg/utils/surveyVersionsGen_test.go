package utils

import (
	"testing"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func TestGenerateSurveyVersionID(t *testing.T) {
	t.Run("test id generation for uniqueness", func(t *testing.T) {
		oldVersions := []*studyTypes.Survey{}

		for i := 0; i < 100; i++ {
			id := GenerateSurveyVersionID(oldVersions)
			oldVersions = append(oldVersions, &studyTypes.Survey{VersionID: id})
		}

		for i, id_1 := range oldVersions {
			for j, id_2 := range oldVersions {
				if i != j && id_1.VersionID == id_2.VersionID {
					t.Errorf("duplicate key present: i: %d - %s j: %d - %s ", i, id_1.VersionID, j, id_2.VersionID)
				}
			}
		}
	})
}
