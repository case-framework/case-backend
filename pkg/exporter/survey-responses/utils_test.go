package surveyresponses

import (
	"testing"

	studytypes "github.com/case-framework/case-backend/pkg/types/study"
)

func TestValueToStr(t *testing.T) {
	tests := []struct {
		name      string
		resultVal interface{}
		expected  string
	}{
		{
			name:      "String",
			resultVal: "hello",
			expected:  "hello",
		},
		{
			name:      "Integer",
			resultVal: 42,
			expected:  "42",
		},
		{
			name:      "Int64",
			resultVal: int64(1234567890),
			expected:  "1234567890",
		},
		{
			name:      "Float64",
			resultVal: 3.14,
			expected:  "3.140000",
		},
		{
			name:      "ResponseItem",
			resultVal: &studytypes.ResponseItem{Key: "1", Value: "answer"},
			expected:  `{"key":"1","value":"answer"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := valueToStr(test.resultVal)
			if result != test.expected {
				t.Errorf("Unexpected result. Got: %s, Expected: %s", result, test.expected)
			}
		})
	}
}

func TestRetrieveResponseItem(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		r := retrieveResponseItem(nil, "")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("retrieve root", func(t *testing.T) {
		r := retrieveResponseItem(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "input"},
				},
			},
		}, "rg")
		if r == nil {
			t.Error("should find result")
		}
	})

	t.Run("retrieve group", func(t *testing.T) {
		r := retrieveResponseItem(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.scg")
		if r == nil {
			t.Error("should find result")
			return
		}
		if r.Key != "scg" || len(r.Items) != 2 {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("retrieve item", func(t *testing.T) {
		r := retrieveResponseItem(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.scg.1")
		if r == nil {
			t.Error("should find result")
			return
		}
		if r.Key != "1" || len(r.Items) != 0 {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("wrong first key", func(t *testing.T) {
		r := retrieveResponseItem(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "wrong.scg.1")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("wrong middle key", func(t *testing.T) {
		r := retrieveResponseItem(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.wrong.1")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("wrong last key", func(t *testing.T) {
		r := retrieveResponseItem(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "rg.scg.wrong")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})
}

func TestRetrieveResponseItemByShortKey(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		r := retrieveResponseItemByShortKey(nil, "")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("retrieve root", func(t *testing.T) {
		r := retrieveResponseItemByShortKey(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "input"},
				},
			},
		}, "rg")
		if r == nil {
			t.Error("should find result")
		}
	})

	t.Run("retrieve group", func(t *testing.T) {
		r := retrieveResponseItemByShortKey(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "scg")
		if r == nil {
			t.Error("should find result")
			return
		}
		if r.Key != "scg" || len(r.Items) != 2 {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("retrieve item", func(t *testing.T) {
		r := retrieveResponseItemByShortKey(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "1")
		if r == nil {
			t.Error("should find result")
			return
		}
		if r.Key != "1" || len(r.Items) != 0 {
			t.Errorf("unexpected result: %v", r)
		}
	})

	t.Run("wrong key", func(t *testing.T) {
		r := retrieveResponseItemByShortKey(&studytypes.SurveyItemResponse{
			Response: &studytypes.ResponseItem{
				Key: "rg",
				Items: []*studytypes.ResponseItem{
					{Key: "scg", Items: []*studytypes.ResponseItem{
						{Key: "1"},
						{Key: "2"},
					}},
				},
			},
		}, "wrong")
		if r != nil {
			t.Errorf("unexpected result: %v", r)
		}
	})

}
