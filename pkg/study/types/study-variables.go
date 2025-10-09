package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

	Key   string             `bson:"key" json:"key"`
	Value any                `bson:"value" json:"value"`
	Type  StudyVariablesType `bson:"type" json:"type"`

	// Metadata for editor
	Label       string `bson:"label" json:"label"`
	Description string `bson:"description" json:"description"`
	UIType      string `bson:"uiType" json:"uiType"`
	UIPriority  int    `bson:"uiPriority" json:"uiPriority"`
	Configs     any    `bson:"configs" json:"configs"`
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

		Key   string             `json:"key"`
		Value json.RawMessage    `json:"value"`
		Type  StudyVariablesType `json:"type"`

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
		// Try number first
		var i int
		if err := json.Unmarshal(raw, &i); err == nil {
			return i, nil
		}
		var f float64
		if err := json.Unmarshal(raw, &f); err == nil {
			// accept 1.0 as 1
			if float64(int(f)) == f {
				return int(f), nil
			}
			return nil, fmt.Errorf("value must be integer, got float: %v", f)
		}
		// Try string formatted integer
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			parsed, perr := strconv.Atoi(s)
			if perr != nil {
				return nil, fmt.Errorf("value string is not integer: %w", perr)
			}
			return parsed, nil
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
