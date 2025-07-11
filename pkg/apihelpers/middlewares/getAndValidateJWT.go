package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

// GetAndValidateJWT is a middleware that extracts the JWT from the request and validates it

func extractAndValidateParticipantJWT(c *gin.Context, tokenSignKey string, globalInfosDBService *globalinfosDB.GlobalInfosDBService) {
	token, err := extractToken(c)
	if err != nil {
		slog.Warn("no Authorization token found")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	if isTokenBlocked(token, globalInfosDBService) {
		slog.Warn("token logged out")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token logged out"})
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
	c.Set("token", token)
	c.Set("validatedToken", parsedToken)
}

func isTokenBlocked(token string, globalInfosDBService *globalinfosDB.GlobalInfosDBService) bool {
	return globalInfosDBService.IsJwtBlocked(token)
}

func GetAndValidateParticipantUserJWT(tokenSignKey string, globalInfosDBService *globalinfosDB.GlobalInfosDBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		extractAndValidateParticipantJWT(c, tokenSignKey, globalInfosDBService)
	}
}

func GetAndValidateParticipantUserJWTWithIgnoringExpiration(tokenSignKey string, globalInfosDBService *globalinfosDB.GlobalInfosDBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c)
		if err != nil {
			slog.Warn("no Authorization token found")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if isTokenBlocked(token, globalInfosDBService) {
			slog.Warn("token logged out")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token logged out"})
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
		c.Set("token", token)
		c.Set("validatedToken", parsedToken)
	}
}

// Can be used by other services to validate the JWT token of the management users
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
