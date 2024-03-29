package apihandlers

import (
	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")

	auth.POST("/send-email",
		mw.HasValidAPIKey(h.apiKeys),
		mw.RequirePayload(),
		h.sendEmail)
}

type HeaderOverrides struct {
	From    string   `json:"from"`
	Sender  string   `json:"sender"`
	ReplyTo []string `json:"replyTo"`
	NoReply bool     `json:"noReply"`
}

type SendEmailReq struct {
	To       []string `json:"to"`
	Subject  string   `json:"subject"`
	Content  string   `json:"content"`
	HighPrio bool     `json:"highPrio"`
}
