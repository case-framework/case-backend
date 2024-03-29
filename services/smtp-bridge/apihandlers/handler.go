package apihandlers

import (
	"net/http"

	sc "github.com/case-framework/case-backend/pkg/smtp-client"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	apiKeys             []string
	highPrioSmtpClients *sc.SmtpClients
	lowPrioSmtpClients  *sc.SmtpClients
}

func NewHTTPHandler(
	apiKeys []string,
	highPrioSmtpClients *sc.SmtpClients,
	lowPrioSmtpClients *sc.SmtpClients,
) *HttpEndpoints {
	return &HttpEndpoints{
		apiKeys:             apiKeys,
		highPrioSmtpClients: highPrioSmtpClients,
		lowPrioSmtpClients:  lowPrioSmtpClients,
	}
}
