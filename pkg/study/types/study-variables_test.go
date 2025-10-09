package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseJSONNumberAsInt64_ZeroAndIntegers(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expected  int64
		wantError bool
	}{
		{name: "zero", input: "0", expected: 0},
		{name: "zero_decimal", input: "0.0", expected: 0},
		{name: "zero_exp", input: "0e0", expected: 0},
		{name: "zero_decimal_exp_shift", input: "0.000e5", expected: 0},
		{name: "plus_zero", input: "+0", expected: 0},
		{name: "minus_zero", input: "-0", expected: 0},
		{name: "leading_zeros_zero", input: "0000.000", expected: 0},

		{name: "int_plain", input: "123", expected: 123},
		{name: "int_with_decimal_point_zeroes", input: "123.000", expected: 123},
		{name: "int_with_leading_zeros", input: "000123", expected: 123},
		{name: "int_with_exp", input: "1e3", expected: 1000},
		{name: "int_with_exp_zero", input: "1e0", expected: 1},
		{name: "negative_with_exp", input: "-2.500e3", expected: -2500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseJSONNumberAsInt64(tc.input)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got value %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestParseJSONNumberAsInt64_NonIntegersRejected(t *testing.T) {
	nonIntegerInputs := []string{"0.001", "1.2", "1e-1", "1.0001", "", ".", "e10", "--1", "1e", "NaN"}
	for _, in := range nonIntegerInputs {
		in := in
		t.Run(in, func(t *testing.T) {
			if _, err := parseJSONNumberAsInt64(in); err == nil {
				t.Fatalf("expected error for input %q, got nil", in)
			}
		})
	}
}

func TestNormalizeStudyVariableValue_IntVariants(t *testing.T) {
	testCases := []struct {
		name      string
		jsonValue string
		expected  int64
	}{
		{name: "number_plain", jsonValue: "1", expected: 1},
		{name: "string_number", jsonValue: "\"1\"", expected: 1},
		{name: "number_decimal_zero", jsonValue: "1.0", expected: 1},
		{name: "string_decimal_zero", jsonValue: "\"1.0\"", expected: 1},
		{name: "exp_number", jsonValue: "1e3", expected: 1000},
		{name: "string_exp_number", jsonValue: "\"1e3\"", expected: 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var raw json.RawMessage = json.RawMessage(tc.jsonValue)
			val, err := normalizeStudyVariableValue(raw, STUDY_VARIABLES_TYPE_INT)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			intVal, ok := val.(int64)
			if !ok {
				t.Fatalf("expected int64, got %T (%v)", val, val)
			}
			if intVal != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, intVal)
			}
		})
	}

	// Non-integer should error
	var raw json.RawMessage = json.RawMessage("0.001")
	if _, err := normalizeStudyVariableValue(raw, STUDY_VARIABLES_TYPE_INT); err == nil {
		t.Fatalf("expected error for non-integer value, got nil")
	}
}

func TestStudyVariables_UnmarshalJSON_Int(t *testing.T) {
	input := []byte(`{
		"type":"int",
		"value":"1.0",
		"studyKey":"study1",
		"key":"k1"
	}`)
	var sv StudyVariables
	if err := sv.UnmarshalJSON(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, ok := sv.Value.(int64)
	if !ok {
		t.Fatalf("expected int64 value, got %T", sv.Value)
	}
	if intVal != 1 {
		t.Fatalf("expected 1, got %d", intVal)
	}
}

func TestStudyVariables_UnmarshalJSON_FloatBoolDate(t *testing.T) {
	// Float
	inputFloat := []byte(`{"type":"float","value":"1.23","studyKey":"s","key":"f"}`)
	var svF StudyVariables
	if err := svF.UnmarshalJSON(inputFloat); err != nil {
		t.Fatalf("float unmarshal error: %v", err)
	}
	if _, ok := svF.Value.(float64); !ok {
		t.Fatalf("expected float64, got %T", svF.Value)
	}

	// Bool
	inputBool := []byte(`{"type":"boolean","value":"true","studyKey":"s","key":"b"}`)
	var svB StudyVariables
	if err := svB.UnmarshalJSON(inputBool); err != nil {
		t.Fatalf("bool unmarshal error: %v", err)
	}
	bVal, ok := svB.Value.(bool)
	if !ok || !bVal {
		t.Fatalf("expected true bool, got %T (%v)", svB.Value, svB.Value)
	}

	// Date RFC3339
	ts := "2024-01-02T03:04:05Z"
	inputDate := []byte(`{"type":"date","value":"` + ts + `","studyKey":"s","key":"d"}`)
	var svD StudyVariables
	if err := svD.UnmarshalJSON(inputDate); err != nil {
		t.Fatalf("date unmarshal error: %v", err)
	}
	tm, ok := svD.Value.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", svD.Value)
	}
	if tm.UTC().Format(time.RFC3339) != ts {
		t.Fatalf("expected %s, got %s", ts, tm.UTC().Format(time.RFC3339))
	}

	// Date Unix seconds (string)
	sec := int64(1704164645)
	inputUnixS := []byte(`{"type":"date","value":"1704164645","studyKey":"s","key":"ds"}`)
	var svDS StudyVariables
	if err := svDS.UnmarshalJSON(inputUnixS); err != nil {
		t.Fatalf("date unix seconds unmarshal error: %v", err)
	}
	tmS, ok := svDS.Value.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time for seconds, got %T", svDS.Value)
	}
	if tmS.Unix() != sec {
		t.Fatalf("expected unix %d, got %d", sec, tmS.Unix())
	}

	// Date Unix millis (string)
	millis := int64(1704164645000)
	inputUnixMs := []byte(`{"type":"date","value":"1704164645000","studyKey":"s","key":"dm"}`)
	var svDM StudyVariables
	if err := svDM.UnmarshalJSON(inputUnixMs); err != nil {
		t.Fatalf("date unix ms unmarshal error: %v", err)
	}
	tmM, ok := svDM.Value.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time for millis, got %T", svDM.Value)
	}
	if tmM.UnixMilli() != millis {
		t.Fatalf("expected unix ms %d, got %d", millis, tmM.UnixMilli())
	}
}

func TestStudyVariables_UnmarshalJSON_IntRejectsNonInteger(t *testing.T) {
	input := []byte(`{"type":"int","value":"0.001","studyKey":"study1","key":"k1"}`)
	var sv StudyVariables
	if err := sv.UnmarshalJSON(input); err == nil {
		t.Fatalf("expected error for non-integer int value, got nil")
	}
}

func TestNormalizeStudyVariableValue_Nulls(t *testing.T) {
	// null or missing decoded as nil
	var rawNull json.RawMessage = json.RawMessage("null")
	val, err := normalizeStudyVariableValue(rawNull, STUDY_VARIABLES_TYPE_INT)
	if err != nil {
		t.Fatalf("unexpected error for null: %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil for null, got %v", val)
	}

	// empty raw should be treated as nil too
	var rawEmpty json.RawMessage
	val, err = normalizeStudyVariableValue(rawEmpty, STUDY_VARIABLES_TYPE_STRING)
	if err != nil {
		t.Fatalf("unexpected error for empty: %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil for empty, got %v", val)
	}
}

func TestStudyVariables_UnmarshalJSON_MetadataRoundtrip(t *testing.T) {
	// Ensure that other fields round-trip without being required for parsing
	input := []byte(`{
		"type":"string",
		"value":"hello",
		"studyKey":"s",
		"key":"greeting",
		"label":"Greeting",
		"description":"A test",
		"uiType":"text",
		"uiPriority":5,
		"configs":{"minLength":1}
	}`)
	var sv StudyVariables
	if err := sv.UnmarshalJSON(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.Type != STUDY_VARIABLES_TYPE_STRING {
		t.Fatalf("expected type string, got %s", sv.Type)
	}
	if sv.Value.(string) != "hello" {
		t.Fatalf("expected value 'hello', got %v", sv.Value)
	}
	if sv.StudyKey != "s" || sv.Key != "greeting" {
		t.Fatalf("unexpected keys: %s/%s", sv.StudyKey, sv.Key)
	}
	if sv.Label != "Greeting" || sv.Description != "A test" || sv.UIType != "text" || sv.UIPriority != 5 {
		t.Fatalf("metadata mismatch")
	}
	// Configs is any; ensure it unmarshals into map
	if _, ok := sv.Configs.(map[string]any); !ok {
		t.Fatalf("configs should be map[string]any, got %T", sv.Configs)
	}
}
