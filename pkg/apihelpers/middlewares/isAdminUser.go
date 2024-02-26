package middlewares

import (
	"log/slog"
	"net/http"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

func IsAdminUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the validated token from the context
		tokenValue, ok := c.Get("validatedToken")
		if !ok {
			slog.Warn("IsAdminUser: validatedToken not found in context")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "validatedToken not found in context"})
			return
		}
		parsedToken := tokenValue.(*jwthandling.ManagementUserClaims)

		if !parsedToken.IsAdmin {
			slog.Warn("IsAdminUser Middleware: non admin user tried to access admin endpoint", slog.String("instanceID", parsedToken.InstanceID), slog.String("userID", parsedToken.Subject))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access to admin endpoint"})
			return
		}
	}
}
