package surveyresponses

import (
	"testing"

	sd "github.com/case-framework/case-backend/pkg/study/exporter/survey-definition"
)

func TestFindSurveyVersion(t *testing.T) {
	t.Run("with no versions available", func(t *testing.T) {
		_, err := findSurveyVersion("id1", 100, []sd.SurveyVersionPreview{})
		if err == nil {
			t.Error("should fail with error")
		}
	})

	testVersions := []sd.SurveyVersionPreview{
		{VersionID: "id1", Published: 0, Unpublished: 50},
		{VersionID: "id2", Published: 50, Unpublished: 120},
		{VersionID: "id3", Published: 120, Unpublished: 0},
	}

	t.Run("with versionID empty - has no matching version based on timestamp", func(t *testing.T) {
		_, err := findSurveyVersion("", -10, testVersions)
		if err == nil {
			t.Error("should fail with error")
		}
	})

	t.Run("with versionID but no matching version", func(t *testing.T) {
		_, err := findSurveyVersion("otherID", -1, testVersions)
		if err == nil {
			t.Error("should fail with error")
		}
	})

	t.Run("with versionID empty - has matching version based on timestamp", func(t *testing.T) {
		sv, err := findSurveyVersion("", 100, testVersions)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if sv.VersionID != "id2" {
			t.Errorf("unexpected version: %v", sv)
		}
	})

	t.Run("with versionID but no matching version but has matching version based on timestamp", func(t *testing.T) {
		sv, err := findSurveyVersion("otherID", 100, testVersions)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if sv.VersionID != "id2" {
			t.Errorf("unexpected version: %v", sv)
		}
	})

	t.Run("with versionID simply", func(t *testing.T) {
		sv, err := findSurveyVersion("id2", 100, testVersions)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if sv.VersionID != "id2" {
			t.Errorf("unexpected version: %v", sv)
		}
	})
}
