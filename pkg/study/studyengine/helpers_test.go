package studyengine

import (
	"testing"
	"time"

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

func TestFormatTimeWithDateFns(t *testing.T) {
	// Test time: December 25, 2023 at 14:30:45 UTC
	testTime := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "Full datetime with 24h format",
			format:   "yyyy-MM-dd HH:mm:ss",
			expected: "2023-12-25 14:30:45",
		},
		{
			name:     "Date only",
			format:   "yyyy-MM-dd",
			expected: "2023-12-25",
		},
		{
			name:     "Time only with 24h format",
			format:   "HH:mm:ss",
			expected: "14:30:45",
		},
		{
			name:     "Time with 12h format and AM/PM",
			format:   "hh:mm:ss a",
			expected: "02:30:45 PM",
		},
		{
			name:     "US date format",
			format:   "MM/dd/yy",
			expected: "12/25/23",
		},
		{
			name:     "European date format",
			format:   "dd.MM.yyyy",
			expected: "25.12.2023",
		},
		{
			name:     "ISO format",
			format:   "yyyy-MM-ddTHH:mm:ss",
			expected: "2023-12-25T14:30:45",
		},
		{
			name:     "Single digit tokens",
			format:   "M/d/yy h:m:s a",
			expected: "12/25/23 2:30:45 PM",
		},
		{
			name:     "Mixed format",
			format:   "yyyy-MM-dd hh:mm a",
			expected: "2023-12-25 02:30 PM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimeWithDateFns(testTime, tt.format)
			if result != tt.expected {
				t.Errorf("FormatTimeWithDateFns() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatTimeWithDateFnsEdgeCases(t *testing.T) {
	// Test edge cases
	tests := []struct {
		name     string
		time     time.Time
		format   string
		expected string
	}{
		{
			name:     "Midnight",
			time:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			format:   "hh:mm a",
			expected: "12:00 AM",
		},
		{
			name:     "Noon",
			time:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			format:   "hh:mm a",
			expected: "12:00 PM",
		},
		{
			name:     "Single digit month and day",
			time:     time.Date(2023, 1, 1, 14, 30, 45, 0, time.UTC),
			format:   "M/d/yyyy",
			expected: "1/1/2023",
		},
		{
			name:     "Milliseconds",
			time:     time.Date(2023, 1, 1, 14, 30, 45, 123000000, time.UTC),
			format:   "HH:mm:ss.SSS",
			expected: "14:30:45.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimeWithDateFns(tt.time, tt.format)
			if result != tt.expected {
				t.Errorf("FormatTimeWithDateFns() = %v, want %v", result, tt.expected)
			}
		})
	}
}
