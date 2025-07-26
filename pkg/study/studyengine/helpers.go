package studyengine

import (
	"errors"
	"fmt"
	"strings"
	"time"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

// Now function control the current time used by the expressions.
var Now func() time.Time = time.Now

// Method to find survey item response in the array of responses
func findSurveyItemResponse(responses []studyTypes.SurveyItemResponse, key string) (responseOfInterest *studyTypes.SurveyItemResponse, err error) {
	for _, response := range responses {
		if response.Key == key {
			return &response, nil
		}
	}
	return nil, errors.New("item not found")
}

// Method to retrive one level of the nested response object
func findResponseObject(surveyItem *studyTypes.SurveyItemResponse, responseKey string) (responseItem *studyTypes.ResponseItem, err error) {
	if surveyItem == nil {
		return responseItem, errors.New("missing survey item")
	}
	if surveyItem.Response == nil {
		return responseItem, errors.New("missing survey item response")
	}
	for i, k := range strings.Split(responseKey, ".") {
		if i == 0 {
			if surveyItem.Response.Key != k {
				// item not found:
				return responseItem, errors.New("response object is not found")
			}
			responseItem = surveyItem.Response
			continue
		}

		found := false
		for _, item := range responseItem.Items {
			if item.Key == k {
				found = true
				responseItem = item
				break
			}
		}
		if !found {
			// item not found:
			return responseItem, errors.New("response object is not found")
		}
	}
	return responseItem, nil
}

func getExternalServicesConfigByName(name string) (ExternalService, error) {
	for _, item := range CurrentStudyEngine.externalServices {
		if item.Name == name {
			return item, nil
		}
	}
	return ExternalService{}, fmt.Errorf("no external service config found with name: %s", name)
}

type ExternalEventPayload struct {
	ParticipantState studyTypes.Participant    `json:"participantState"`
	EventType        string                    `json:"eventType"`
	StudyKey         string                    `json:"studyKey"`
	InstanceID       string                    `json:"instanceID"`
	Response         studyTypes.SurveyResponse `json:"surveyResponses"`
	EventKey         string                    `json:"eventKey"`
	Payload          map[string]interface{}    `json:"payload"`
}

// Helper function to convert a date to a string

// dateFnsTokenMap maps date-fns style tokens to Go's layout string equivalents
var dateFnsTokenMap = map[string]string{
	// Year tokens
	"yyyy": "2006", // 4-digit year
	"yy":   "06",   // 2-digit year

	// Month tokens
	"MM": "01", // 2-digit month (01-12)
	"M":  "1",  // 1-digit month (1-12)

	// Day tokens
	"dd": "02", // 2-digit day (01-31)
	"d":  "2",  // 1-digit day (1-31)

	// Hour tokens (24-hour format)
	"HH": "15", // 2-digit hour (00-23)

	// Hour tokens (12-hour format)
	"hh": "03", // 2-digit hour (01-12)

	// Minute tokens
	"mm": "04", // 2-digit minute (00-59)
	"m":  "4",  // 1-digit minute (0-59)

	// Second tokens
	"ss": "05", // 2-digit second (00-59)
	"s":  "5",  // 1-digit second (0-59)

	// AM/PM tokens
	"a":  "PM", // AM/PM
	"aa": "PM", // AM/PM (alternative)
}

// FormatTimeWithDateFns formats a time.Time value using date-fns style tokens
//
// Usage example:
//
//	t := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)
//	formatted := FormatTimeWithDateFns(t, "yyyy-MM-dd HH:mm:ss")
//	// Result: "2023-12-25 14:30:45"
//
//	formatted2 := FormatTimeWithDateFns(t, "MM/dd/yy hh:mm a")
//	// Result: "12/25/23 02:30 PM"
//
// Supported tokens:
//   - yyyy: 4-digit year (2006)
//   - yy: 2-digit year (06)
//   - MM: 2-digit month (01-12)
//   - M: 1-digit month (1-12)
//   - dd: 2-digit day (01-31)
//   - d: 1-digit day (1-31)
//   - HH: 2-digit hour 24h (00-23)
//   - hh: 2-digit hour 12h (01-12)
//   - mm: 2-digit minute (00-59)
//   - m: 1-digit minute (0-59)
//   - ss: 2-digit second (00-59)
//   - s: 1-digit second (0-59)
//   - a: AM/PM indicator
func FormatTimeWithDateFns(t time.Time, format string) string {
	result := format

	// Replace tokens in order of length (longest first) to avoid conflicts
	tokens := []string{
		"yyyy", "yy",
		"SSS", "SS", "S",
		"MM", "M",
		"dd", "d",
		"HH", "H",
		"hh", "h",
		"mm", "m",
		"ss", "s",
		"aa", "a",
	}

	// Process predefined tokens first
	for _, token := range tokens {
		if goLayout, exists := dateFnsTokenMap[token]; exists {
			result = strings.ReplaceAll(result, token, goLayout)
		}
	}

	// Process any custom tokens that might have been added
	for token, goLayout := range dateFnsTokenMap {
		// Skip predefined tokens that were already processed
		found := false
		for _, predefinedToken := range tokens {
			if token == predefinedToken {
				found = true
				break
			}
		}
		if !found {
			result = strings.ReplaceAll(result, token, goLayout)
		}
	}

	return t.Format(result)
}
