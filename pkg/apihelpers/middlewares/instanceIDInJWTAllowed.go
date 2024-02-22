package middlewares

import (
	"log/slog"
	"net/http"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

// GetAndValidateJWT is a middleware that extracts the JWT from the request and validates it
func IsInstanceIDInJWTAllowed(allowedInstanceIDs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the validated token from the context
		parsedToken, ok := c.Get("validatedToken")
		if !ok {
			slog.Warn("IsInstanceIDInJWTAllowed Middleware: validatedToken not found in context")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "validatedToken not found in context"})
			return
		}

		// Check if the instanceID in the token is allowed
		instanceID := parsedToken.(*jwthandling.ManagementUserClaims).InstanceID

		// Check if the instanceID is allowed
		allowed := false
		for _, allowedInstanceID := range allowedInstanceIDs {
			if instanceID == allowedInstanceID {
				allowed = true
				break
			}
		}

		if !allowed {
			slog.Warn("IsInstanceIDInJWTAllowed Middleware: instanceID not allowed", slog.String("instanceID", instanceID))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "instanceID not allowed"})
			return
		}
	}
}
