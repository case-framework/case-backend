package apihandlers

import (
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	sc "github.com/case-framework/case-backend/pkg/smtp-client"

	"github.com/gin-gonic/gin"
)

const (
	maxRetry = 5
)

func (h *HttpEndpoints) AddRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/")

	auth.POST("/send-email",
		mw.HasValidAPIKey(h.apiKeys),
		mw.RequirePayload(),
		h.sendEmail)
}

type SendEmailReq struct {
	To              []string            `json:"to"`
	Subject         string              `json:"subject"`
	Content         string              `json:"content"`
	HighPrio        bool                `json:"highPrio"`
	HeaderOverrides *sc.HeaderOverrides `json:"headerOverrides"`
}

func (h *HttpEndpoints) sendEmail(c *gin.Context) {
	var req SendEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.To) < 1 {
		slog.Error("missing 'to' field")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'to' field"})
		return
	}

	retryCounter := 0
	for {
		var err error
		if req.HighPrio {
			err = h.highPrioSmtpClients.SendMail(
				req.To,
				req.Subject,
				req.Content,
				req.HeaderOverrides,
			)
		} else {
			err = h.lowPrioSmtpClients.SendMail(
				req.To,
				req.Subject,
				req.Content,
				req.HeaderOverrides,
			)
		}
		if err != nil {
			if retryCounter >= maxRetry {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email"})
				return
			}
			retryCounter += 1
			slog.Error("failed to send email", slog.String("error", err.Error()), slog.Int("retryCounter", retryCounter))
			time.Sleep(time.Duration(retryCounter) * time.Second)
		} else {
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "email sent"})
}
