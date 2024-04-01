package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddParticipantAuthAPI(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")

	authGroup.POST("/login", mw.RequirePayload(), h.loginWithEmail)

}

type LoginWithEmailReq struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	InstanceID string `json:"instanceId"`
}

func (h *HttpEndpoints) loginWithEmail(c *gin.Context) {
	var req LoginWithEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: check request
	// ...

}
