package studyutils

import (
	"maps"
	"slices"
	"strings"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

type ConfidentialResponsesExportEntry struct {
	ParticipantID string `json:"participantID"`
	EntryID       string `json:"entryID"`
	ResponseKey   string `json:"responseKey"`
	Value         string `json:"value"`
}

func PrepConfidentialResponseExport(resp studyTypes.SurveyResponse, realPID string, respKeyFilter []string) []ConfidentialResponsesExportEntry {
	parsedResp := []ConfidentialResponsesExportEntry{}

	for _, r := range resp.Responses {
		slotKey := r.Key + "-"

		slots := parseSlots(r.Response, slotKey)
		for k, v := range slots {
			if len(respKeyFilter) > 0 && !slices.Contains(respKeyFilter, k) {
				continue
			}
			parsedResp = append(parsedResp, ConfidentialResponsesExportEntry{
				ParticipantID: realPID,
				EntryID:       resp.ID.Hex(),
				ResponseKey:   k,
				Value:         v,
			})
		}
	}

	return parsedResp
}

func parseSlots(respItem *studyTypes.ResponseItem, slotKey string) map[string]string {
	parsedResp := map[string]string{}
	if respItem == nil {
		return parsedResp
	}

	currentSlotKey := slotKey + "." + respItem.Key
	if strings.HasSuffix(slotKey, "-") {
		currentSlotKey = slotKey + respItem.Key
	}

	if len(respItem.Items) == 0 {
		parsedResp[currentSlotKey] = respItem.Value
		return parsedResp
	}

	for _, subItem := range respItem.Items {
		r := parseSlots(subItem, currentSlotKey)
		maps.Copy(parsedResp, r)
	}
	return parsedResp
}
