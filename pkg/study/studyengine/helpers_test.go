package studyengine

import (
	"testing"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func TestFindSurveyItemResponse(t *testing.T) {

	t.Run("empty array", func(t *testing.T) {
		_, err := findSurveyItemResponse([]studyTypes.SurveyItemResponse{}, "t.G1.4")
		if err == nil {
			t.Error("should produce error")
		}
	})
	t.Run("key not present", func(t *testing.T) {
		_, err := findSurveyItemResponse([]studyTypes.SurveyItemResponse{
			{Key: "t.G1.1"},
			{Key: "t.G1.2"},
			{Key: "t.G1.3"},
			{Key: "t.G2.1"},
		}, "t.G1.4")
		if err == nil {
			t.Error("should produce error")
		}
	})
	t.Run("key present", func(t *testing.T) {
		item, err := findSurveyItemResponse([]studyTypes.SurveyItemResponse{
			{Key: "t.G1.1"},
			{Key: "t.G1.2"},
			{Key: "t.G1.3"},
			{Key: "t.G1.4"},
			{Key: "t.G2.1"},
		}, "t.G1.4")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if item.Key != "t.G1.4" {
			t.Errorf("unexpected item: %v", item)
		}
	})
}

func TestFindResponseObject(t *testing.T) {

	t.Run("nil item", func(t *testing.T) {
		_, err := findResponseObject(nil, "rg.scg.1")
		if err == nil {
			t.Error("should produce error")
		}

	})

	t.Run("no responses", func(t *testing.T) {
		_, err := findResponseObject(&studyTypes.SurveyItemResponse{
			Key: "test",
		}, "rg.scg.1")
		if err == nil {
			t.Error("should produce error")
		}
	})

	t.Run("response parent missing", func(t *testing.T) {
		_, err := findResponseObject(&studyTypes.SurveyItemResponse{
			Key: "test",
			Response: &studyTypes.ResponseItem{
				Key: "rgwrong",
				Items: []*studyTypes.ResponseItem{
					{Key: "scg", Items: []*studyTypes.ResponseItem{
						{Key: "1"},
					}},
				},
			},
		}, "rg.scg.1")
		if err == nil {
			t.Error("should produce error")
		}

	})

	t.Run("final response missing", func(t *testing.T) {
		_, err := findResponseObject(&studyTypes.SurveyItemResponse{
			Key: "test",
			Response: &studyTypes.ResponseItem{
				Key: "rg",
				Items: []*studyTypes.ResponseItem{
					{Key: "scg", Items: []*studyTypes.ResponseItem{
						{Key: "2"},
					}},
				},
			},
		}, "rg.scg.1")
		if err == nil {
			t.Error("should produce error")
		}
	})

	t.Run("response correct", func(t *testing.T) {
		response, err := findResponseObject(&studyTypes.SurveyItemResponse{
			Key: "test",
			Response: &studyTypes.ResponseItem{
				Key: "rg",
				Items: []*studyTypes.ResponseItem{
					{Key: "scg", Items: []*studyTypes.ResponseItem{
						{Key: "1", Value: "testvalue"},
					}},
				},
			},
		}, "rg.scg.1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if response.Value != "testvalue" {
			t.Errorf("unexpected item: %v", response)
		}
	})
}

func TestGetExternalServicesConfigByName(t *testing.T) {
	configs := []ExternalService{
		{Name: "test1", URL: "url1"},
		{Name: "test2", URL: "url2"},
	}

	CurrentStudyEngine.externalServices = configs

	t.Run("item not there", func(t *testing.T) {
		_, err := getExternalServicesConfigByName("wrong")
		if err == nil {
			t.Error("should produce error")
		}
	})

	t.Run("item there", func(t *testing.T) {
		conf, err := getExternalServicesConfigByName("test2")
		if err != nil {
			t.Errorf("unexpected error %v", err)
			return
		}
		if conf.URL != "url2" {
			t.Errorf("unexpected values %v", conf)
		}
	})
}
