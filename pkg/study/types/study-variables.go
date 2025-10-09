package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyVariablesType string

const (
	STUDY_VARIABLES_TYPE_STRING  StudyVariablesType = "string"
	STUDY_VARIABLES_TYPE_INT     StudyVariablesType = "int"
	STUDY_VARIABLES_TYPE_FLOAT   StudyVariablesType = "float"
	STUDY_VARIABLES_TYPE_BOOLEAN StudyVariablesType = "boolean"
	STUDY_VARIABLES_TYPE_DATE    StudyVariablesType = "date"
)

type StudyVariables struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	ConfigUpdatedAt time.Time          `bson:"configUpdatedAt" json:"configUpdatedAt"`
	ValueUpdatedAt  time.Time          `bson:"valueUpdatedAt" json:"valueUpdatedAt"`

	StudyKey string `bson:"studyKey" json:"studyKey"`

	Key   string             `bson:"key" json:"key"`
	Value any                `bson:"value" json:"value"`
	Type  StudyVariablesType `bson:"type" json:"type"`

	// Metadata for editor
	Label       string `bson:"label" json:"label,omitempty"`
	Description string `bson:"description" json:"description,omitempty"`
	UIType      string `bson:"uiType" json:"uiType,omitempty"`
	UIPriority  int    `bson:"uiPriority" json:"uiPriority,omitempty"`
	Configs     any    `bson:"configs" json:"configs,omitempty"`
}

// UnmarshalJSON normalizes the Value field to the correct Go type based on Type.
// This ensures we persist correct BSON types and round-trip cleanly back to JSON.
func (sv *StudyVariables) UnmarshalJSON(data []byte) error {
	// Define a wire struct that treats Value as raw JSON.
	type studyVariablesWire struct {
		ID              primitive.ObjectID `json:"id,omitempty"`
		CreatedAt       time.Time          `json:"createdAt"`
		ConfigUpdatedAt time.Time          `json:"configUpdatedAt"`
		ValueUpdatedAt  time.Time          `json:"valueUpdatedAt"`

		StudyKey string             `json:"studyKey"`
		Key      string             `json:"key"`
		Value    json.RawMessage    `json:"value"`
		Type     StudyVariablesType `json:"type"`

		Label       string `json:"label"`
		Description string `json:"description"`
		UIType      string `json:"uiType"`
		UIPriority  int    `json:"uiPriority"`
		Configs     any    `json:"configs"`
	}

	var wire studyVariablesWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	normalizedValue, err := normalizeStudyVariableValue(wire.Value, wire.Type)
	if err != nil {
		return err
	}

	sv.ID = wire.ID
	sv.CreatedAt = wire.CreatedAt
	sv.ConfigUpdatedAt = wire.ConfigUpdatedAt
	sv.ValueUpdatedAt = wire.ValueUpdatedAt
	sv.StudyKey = wire.StudyKey
	sv.Key = wire.Key
	sv.Value = normalizedValue
	sv.Type = wire.Type
	sv.Label = wire.Label
	sv.Description = wire.Description
	sv.UIType = wire.UIType
	sv.UIPriority = wire.UIPriority
	sv.Configs = wire.Configs
	return nil
}

func normalizeStudyVariableValue(raw json.RawMessage, t StudyVariablesType) (any, error) {
	// Treat missing or explicit null as nil value
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	switch t {
	case STUDY_VARIABLES_TYPE_STRING:
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("value must be string: %w", err)
		}
		return s, nil

	case STUDY_VARIABLES_TYPE_INT:
		// Accept integers as JSON numbers or numeric strings without losing precision.
		// 1) If it's a quoted string, parse as integer (supports forms like "1", "1.0", "1e3").
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			if parsed, perr := strconv.ParseInt(s, 10, 64); perr == nil {
				return parsed, nil
			}
			if i64, perr := parseJSONNumberAsInt64(s); perr == nil {
				return i64, nil
			}
			return nil, fmt.Errorf("value string is not integer")
		}

		// 2) Parse unquoted JSON number using json.Number to avoid float64 precision loss.
		var num json.Number
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()
		if err := dec.Decode(&num); err == nil {
			if i64, ierr := num.Int64(); ierr == nil {
				return i64, nil
			}
			if i64, ierr := parseJSONNumberAsInt64(num.String()); ierr == nil {
				return i64, nil
			}
			return nil, fmt.Errorf("value must be integer, got number: %s", num.String())
		}
		return nil, errors.New("value must be integer")

	case STUDY_VARIABLES_TYPE_FLOAT:
		var f float64
		if err := json.Unmarshal(raw, &f); err == nil {
			return f, nil
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			parsed, perr := strconv.ParseFloat(s, 64)
			if perr != nil {
				return nil, fmt.Errorf("value string is not float: %w", perr)
			}
			return parsed, nil
		}
		return nil, errors.New("value must be float")

	case STUDY_VARIABLES_TYPE_BOOLEAN:
		var b bool
		if err := json.Unmarshal(raw, &b); err == nil {
			return b, nil
		}
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			parsed, perr := strconv.ParseBool(s)
			if perr != nil {
				return nil, fmt.Errorf("value string is not boolean: %w", perr)
			}
			return parsed, nil
		}
		return nil, errors.New("value must be boolean")

	case STUDY_VARIABLES_TYPE_DATE:
		// Accept RFC3339 string, RFC3339Nano string, or unix seconds/milliseconds
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			if s == "" {
				return time.Time{}, nil
			}
			if ts, perr := strconv.ParseInt(s, 10, 64); perr == nil {
				// Numeric string, interpret as unix seconds or ms
				return unixToTime(ts), nil
			}
			// Try common time formats
			if tVal, perr := time.Parse(time.RFC3339Nano, s); perr == nil {
				return tVal, nil
			}
			if tVal, perr := time.Parse(time.RFC3339, s); perr == nil {
				return tVal, nil
			}
			return nil, fmt.Errorf("invalid RFC3339 date string: %s", s)
		}
		var n int64
		if err := json.Unmarshal(raw, &n); err == nil {
			return unixToTime(n), nil
		}
		var fn float64
		if err := json.Unmarshal(raw, &fn); err == nil {
			return unixToTime(int64(fn)), nil
		}
		return nil, errors.New("value must be date as RFC3339 string or unix timestamp")
	}

	// Unknown type
	return nil, fmt.Errorf("unsupported study variable type: %s", t)
}

func unixToTime(v int64) time.Time {
	// Heuristic: treat >= 1e12 as milliseconds since epoch, else seconds
	if v >= 1_000_000_000_000 {
		sec := v / 1000
		nsec := (v % 1000) * int64(time.Millisecond)
		return time.Unix(sec, nsec).UTC()
	}
	return time.Unix(v, 0).UTC()
}

// parseJSONNumberAsInt64 parses a JSON number string (which may include a decimal
// point or exponent) and returns an int64 if and only if the numeric value is an
// exact integer that fits in the int64 range. Examples accepted: "1", "1.0",
// "1e3", "-2.500e3" (equals -2500). Non-integer values are rejected.
func parseJSONNumberAsInt64(numStr string) (int64, error) {
	s := strings.TrimSpace(numStr)
	if s == "" {
		return 0, errors.New("empty number")
	}

	// Extract sign
	negative := false
	if s[0] == '+' {
		s = s[1:]
	} else if s[0] == '-' {
		negative = true
		s = s[1:]
	}
	if s == "" {
		return 0, errors.New("invalid number")
	}

	// Split exponent if present
	var exp int64
	if idx := strings.IndexAny(s, "eE"); idx != -1 {
		expPart := s[idx+1:]
		s = s[:idx]
		if expPart == "" {
			return 0, errors.New("invalid exponent")
		}
		e, err := strconv.ParseInt(expPart, 10, 64)
		if err != nil {
			return 0, errors.New("invalid exponent")
		}
		exp = e
		if s == "" {
			// No mantissa before exponent, e.g., "e10"
			return 0, errors.New("invalid number")
		}
	}

	// Split integer and fractional parts
	intPart := s
	fracPart := ""
	hadDot := false
	if dot := strings.IndexByte(s, '.'); dot != -1 {
		hadDot = true
		intPart = s[:dot]
		fracPart = s[dot+1:]
	}
	if intPart == "" {
		intPart = "0"
	}
	// Reject formats with no digits around a dot (".") or trailing dot ("1.")
	if hadDot && (len(intPart) == 0 && len(fracPart) == 0) {
		return 0, errors.New("invalid number")
	}
	if hadDot && len(fracPart) == 0 && len(intPart) > 0 {
		return 0, errors.New("invalid number")
	}

	// Validate digits-only
	for _, ch := range intPart {
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid digits")
		}
	}
	for _, ch := range fracPart {
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid digits")
		}
	}

	// Fast-path zero: if both parts are all zeros, value is zero regardless of exponent
	isAllZeros := func(str string) bool {
		for _, ch := range str {
			if ch != '0' {
				return false
			}
		}
		return len(str) > 0
	}
	if isAllZeros(intPart) && (fracPart == "" || isAllZeros(fracPart)) {
		return 0, nil
	}

	// Work on the full digit sequence without trimming leading zeros yet
	digits := intPart + fracPart

	// Positive scale means digits still contain fractional information after exponent shift
	scale := int64(len(fracPart)) - exp
	if scale > 0 {
		// Integer only if the last 'scale' digits are zeros
		if int(scale) > len(digits) {
			// Only integer if all existing digits are zeros (i.e., numeric value equals 0)
			allZero := true
			for _, ch := range digits {
				if ch != '0' {
					allZero = false
					break
				}
			}
			if !allZero {
				return 0, errors.New("not integer")
			}
			return 0, nil
		}
		for _, ch := range digits[len(digits)-int(scale):] {
			if ch != '0' {
				return 0, errors.New("not integer")
			}
		}
		// Remove the fractional zeros
		digits = digits[:len(digits)-int(scale)]
	} else if scale < 0 {
		// Append zeros to shift decimal to the right
		digits += strings.Repeat("0", int(-scale))
	}

	// Now trim leading zeros to minimize the integer representation size
	i := 0
	for i < len(digits) && digits[i] == '0' {
		i++
	}
	digits = digits[i:]

	if digits == "" {
		return 0, nil
	}

	bi := new(big.Int)
	if _, ok := bi.SetString(digits, 10); !ok {
		return 0, errors.New("invalid integer")
	}
	if negative {
		bi.Neg(bi)
	}

	if !bi.IsInt64() {
		return 0, errors.New("integer out of range")
	}
	return bi.Int64(), nil
}

type StudyVariableIntConfig struct {
	Min int `bson:"min" json:"min"`
	Max int `bson:"max" json:"max"`
}

type StudyVariableFloatConfig struct {
	Min float64 `bson:"min" json:"min"`
	Max float64 `bson:"max" json:"max"`
}

type StudyVariableStringConfig struct {
	MinLength      int      `bson:"minLength" json:"minLength"`
	MaxLength      int      `bson:"maxLength" json:"maxLength"`
	Pattern        string   `bson:"pattern" json:"pattern"`
	PossibleValues []string `bson:"possibleValues" json:"possibleValues"`
}

type StudyVariableDateConfig struct {
	Min time.Time `bson:"min" json:"min"`
	Max time.Time `bson:"max" json:"max"`
}
