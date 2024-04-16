package apihandlers

import (
	"log/slog"
	"net/http"
	"strings"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	pc "github.com/case-framework/case-backend/pkg/permission-checker"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) isInstanceAllowed(instanceID string) bool {
	for _, id := range h.allowedInstanceIDs {
		if id == instanceID {
			return true
		}
	}
	return false
}

type RequiredPermission struct {
	ResourceType        string
	ResourceKeys        []string
	ExtractResourceKeys func(c *gin.Context) []string // from route params, query params, etc.
	Action              string
}

func (h *HttpEndpoints) useAuthorisedHandler(
	requiredPermission RequiredPermission,
	getLimiterRequirement func(c *gin.Context) map[string]string,
	handler gin.HandlerFunc,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

		var limiterReq map[string]string
		if getLimiterRequirement != nil {
			limiterReq = getLimiterRequirement(c)
		}

		rks := requiredPermission.ResourceKeys
		if requiredPermission.ExtractResourceKeys != nil {
			newRks := requiredPermission.ExtractResourceKeys(c)
			rks = append(rks, newRks...)
		}

		hasPermission := pc.IsAuthorized(
			h.muDBConn,
			token.IsAdmin,
			token.InstanceID,
			token.Subject,
			pc.SUBJECT_TYPE_MANAGEMENT_USER,
			requiredPermission.ResourceType,
			rks,
			requiredPermission.Action,
			limiterReq,
		)
		if !hasPermission {
			slog.Warn("unauthorised access attempted",
				slog.String("instanceID", token.InstanceID),
				slog.String("userID", token.Subject),
				slog.String("resourceType", requiredPermission.ResourceType),
				slog.String("resourceKeys", strings.Join(rks, ",")),
				slog.String("action", requiredPermission.Action),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorised access attempted"})
			return
		}

		handler(c)
	}
}

func (h *HttpEndpoints) onExportTaskFailed(
	instanceID string,
	taskID string,
	errMsg string,
) {
	err := h.studyDBConn.UpdateTaskCompleted(
		instanceID,
		taskID,
		studyTypes.TASK_STATUS_COMPLETED,
		0,
		errMsg,
		"",
	)
	if err != nil {
		slog.Error("failed to update task status", slog.String("error", err.Error()), slog.String("taskID", taskID))
	}
}
