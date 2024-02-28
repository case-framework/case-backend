package apihandlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	pc "github.com/case-framework/case-backend/pkg/permission-checker"
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
	ResourceType string
	ResourceKeys []string
	Action       string
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

		hasPermission := pc.IsAuthorized(
			h.muDBConn,
			token.IsAdmin,
			token.InstanceID,
			token.Subject,
			pc.SUBJECT_TYPE_MANAGEMENT_USER,
			requiredPermission.ResourceType,
			requiredPermission.ResourceKeys,
			requiredPermission.Action,
			limiterReq,
		)
		if !hasPermission {
			requiredPermissionStr, _ := json.Marshal(requiredPermission)
			slog.Warn("unauthorised access attempted", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requiredPermission", string(requiredPermissionStr)))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorised access attempted"})
			return
		}

		handler(c)
	}
}
