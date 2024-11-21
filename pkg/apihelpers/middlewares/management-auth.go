package middlewares

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	mudb "github.com/case-framework/case-backend/pkg/db/management-user"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderAPIKey        = "X-API-Key"
	HeaderInstanceID    = "X-Instance-ID"
)

func ManagementAuthMiddleware(tokenSignKey string, allowedInstanceIds []string, muDB *mudb.ManagementUserDBService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isServiceUser(c) {
			validateServiceUser(c, allowedInstanceIds, muDB)
		} else {
			validateManagementUser(c, tokenSignKey, allowedInstanceIds)
		}
	}
}

func isServiceUser(c *gin.Context) bool {
	apiKey := c.GetHeader(HeaderAPIKey)
	instanceID := c.GetHeader(HeaderInstanceID)
	if apiKey != "" && instanceID != "" {
		return true
	}
	return false
}

func validateServiceUser(c *gin.Context, allowedInstanceIds []string, muDB *mudb.ManagementUserDBService) {
	slog.Debug("auth as service user")
	apiKey := c.GetHeader(HeaderAPIKey)
	instanceID := c.GetHeader(HeaderInstanceID)

	if !isInstanceAllowed(instanceID, allowedInstanceIds) {
		slog.Warn("instanceID not allowed", slog.String("instanceID", instanceID), slog.String("path", c.Request.URL.Path))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "instanceID not allowed"})
		c.Abort()
		return
	}

	sApiKey, err := muDB.GetServiceUserAPIKey(instanceID, apiKey)
	if err != nil {
		slog.Warn("Attempted to use invalid api key", slog.String("apiKey", apiKey), slog.String("path", c.Request.URL.Path))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		c.Abort()
		return
	}

	if sApiKey.ExpiresAt.Before(time.Now()) {
		slog.Warn("Attempted to use expired api key", slog.String("apiKey", apiKey), slog.String("path", c.Request.URL.Path))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "expired api key"})
		c.Abort()
		return
	}

	parsedToken := &jwthandling.ManagementUserClaims{
		InstanceID:    instanceID,
		IsAdmin:       false,
		IsServiceUser: true,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: sApiKey.ServiceUserID,
		},
	}
	c.Set("validatedToken", parsedToken)
	c.Next()

}

func validateManagementUser(c *gin.Context, tokenSignKey string, allowedInstanceIDs []string) {
	slog.Debug("auth as management user")
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

	// Check if the instanceID is allowed
	if !isInstanceAllowed(parsedToken.InstanceID, allowedInstanceIDs) {
		slog.Warn("instanceID not allowed", slog.String("instanceID", parsedToken.InstanceID), slog.String("path", c.Request.URL.Path))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "instanceID not allowed"})
		c.Abort()
		return
	}
	c.Set("validatedToken", parsedToken)
}

func extractToken(c *gin.Context) (string, error) {
	req := c.Request

	var token string
	tokens, ok := req.Header[HeaderAuthorization]
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

func isInstanceAllowed(instanceID string, allowedInstanceIDs []string) bool {
	for _, id := range allowedInstanceIDs {
		if id == instanceID {
			return true
		}
	}
	return false
}
