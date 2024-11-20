package apihandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	apiKeys   []string
	emailsDir string
}

func NewHTTPHandler(apiKeys []string, emailsDir string) *HttpEndpoints {
	return &HttpEndpoints{
		apiKeys:   apiKeys,
		emailsDir: emailsDir,
	}
}
