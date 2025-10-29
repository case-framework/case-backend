package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func HasValidAPIKey(validKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request

		keysInHeader, ok := req.Header["Api-Key"]
		if !ok || len(keysInHeader) < 1 {
			slog.Error("A valid API key missing")
			c.JSON(http.StatusBadRequest, gin.H{"error": "A valid API key missing"})
			c.Abort()
			return
		}

		for _, k := range keysInHeader {
			for _, vk := range validKeys {
				if k == vk {
					c.Next()
					return
				}
			}
		}

		// If no keys matched:
		slog.Error("A valid API key missing")
		slog.Debug("Received API keys", slog.String("receivedKeys", strings.Join(keysInHeader, ",")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid API key missing"})
		c.Abort()
	}
}
