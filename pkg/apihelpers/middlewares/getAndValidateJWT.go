package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

// GetAndValidateJWT is a middleware that extracts the JWT from the request and validates it

func extractAndValidateParticipantJWT(c *gin.Context, tokenSignKey string) {
	token, err := extractToken(c)
	if err != nil {
		slog.Warn("no Authorization token found")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Parse and validate token
	parsedToken, ok, err := jwthandling.ValidateParticipantUserToken(token, tokenSignKey)
	if err != nil || !ok {
		slog.Warn("token validation failed", slog.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "error during token validation"})
		c.Abort()
		return
	}
	c.Set("validatedToken", parsedToken)
}

func GetAndValidateParticipantUserJWT(tokenSignKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		extractAndValidateParticipantJWT(c, tokenSignKey)
	}
}

func GetAndValidateParticipantUserJWTWithIgnoringExpiration(tokenSignKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c)
		if err != nil {
			slog.Warn("no Authorization token found")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Parse and validate token
		parsedToken, _, err := jwthandling.ValidateParticipantUserToken(token, tokenSignKey)
		if err != nil && !strings.Contains(err.Error(), "token is expired") {
			slog.Warn("token validation failed", slog.String("error", err.Error()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "error during token validation"})
			c.Abort()
			return
		}
		c.Set("validatedToken", parsedToken)
	}
}
