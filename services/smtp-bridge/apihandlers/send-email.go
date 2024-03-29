package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	sc "github.com/case-framework/case-backend/pkg/smtp-client"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")

	auth.POST("/send-email",
		mw.HasValidAPIKey(h.apiKeys),
		mw.RequirePayload(),
		h.sendEmail)
}

type SendEmailReq struct {
	To              []string           `json:"to"`
	Subject         string             `json:"subject"`
	Content         string             `json:"content"`
	HighPrio        bool               `json:"highPrio"`
	HeaderOverrides sc.HeaderOverrides `json:"headerOverrides"`
}

func (h *HttpEndpoints) sendEmail(c *gin.Context) {
	var req SendEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

}
