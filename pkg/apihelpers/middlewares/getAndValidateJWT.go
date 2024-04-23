package middlewares

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

// GetAndValidateJWT is a middleware that extracts the JWT from the request and validates it
func GetAndValidateManagementUserJWT(tokenSignKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c)
		if err != nil {
			slog.Warn("no Authorization token found")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Parse and validate token
		parsedToken, ok, err := jwthandling.ValidateManagementUserToken(token, tokenSignKey)
		if err != nil || !ok {
			slog.Warn("token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "error during token validation"})
			c.Abort()
			return
		}
		c.Set("validatedToken", parsedToken)
	}
}

func GetAndValidateParticipantUserJWT(tokenSignKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			slog.Warn("token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "error during token validation"})
			c.Abort()
			return
		}
		c.Set("validatedToken", parsedToken)
	}
}

func extractToken(c *gin.Context) (string, error) {
	req := c.Request

	var token string
	tokens, ok := req.Header["Authorization"]
	if ok && len(tokens) > 0 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
		if len(token) == 0 {
			return token, errors.New("No token found in Authorization header")
		}
	} else {
		return token, errors.New("No Authorization header found")
	}
	return token, nil
}
