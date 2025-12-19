package study

import (
	"log/slog"

	"github.com/case-framework/case-backend/pkg/study/studyengine"
)

func IsAllowedToUploadFile(
	instanceID string,
	studyKey string,
	profileID string,
	fileSize int64,
	contentType string,
) (allowed bool, forParticipantID string) {
	// Get study to compute participantID
	study, err := studyDBService.GetStudy(instanceID, studyKey)
	if err != nil {
		slog.Error("failed to get study", slog.String("error", err.Error()))
		return false, ""
	}

	// Check if file upload is allowed for this study
	// If no rule is configured, deny by default
	if study.Configs.ParticipantFileUploadRule == nil {
		slog.Warn("File upload not allowed - no rule configured", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("profileID", profileID))
		return false, ""
	}

	// Compute participantID from profileID
	participantID, _, err := ComputeParticipantIDs(study, profileID)
	if err != nil {
		slog.Error("Error computing participant IDs", slog.String("instanceID", instanceID), slog.String("studyKey", study.Key), slog.String("error", err.Error()))
		return false, ""
	}

	// Get participant state for evaluation context
	pState, err := studyDBService.GetParticipantByID(instanceID, studyKey, participantID)
	if err != nil {
		// If participant doesn't exist yet, create a minimal state for evaluation
		slog.Error("participant not found", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("participantID", participantID))
		return false, ""
	}

	// Create evaluation context
	evalCtx := studyengine.EvalContext{
		Event: studyengine.StudyEvent{
			InstanceID: instanceID,
			StudyKey:   studyKey,
			Type:       studyengine.STUDY_EVENT_TYPE_CUSTOM,
			EventKey:   "FILE_UPLOAD",
			Payload:    map[string]any{"fileSize": fileSize, "contentType": contentType},
		},
		ParticipantState: pState,
	}

	// Evaluate the file upload rule
	result, err := studyengine.ExpressionEval(*study.Configs.ParticipantFileUploadRule, evalCtx)
	if err != nil {
		slog.Error("Error evaluating file upload rule", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("error", err.Error()))
		return false, ""
	}

	// Check if upload is allowed (rule must evaluate to true)
	allowed, ok := result.(bool)
	if !ok {
		// Try to convert numeric result to bool (non-zero = true)
		if numVal, ok := result.(float64); ok {
			allowed = numVal != 0
		} else {
			slog.Error("File upload rule returned unexpected type", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
			return false, ""
		}
	}

	if !allowed {
		slog.Warn("File upload not allowed by study rule", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("profileID", profileID))
		return false, ""
	}

	return true, participantID
}
