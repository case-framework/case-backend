package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddManagementAuthAPI(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/signin-with-idp", mw.RequirePayload(), h.signInWithIdP)
	auth.GET("/renew-token", mw.GetAndValidateManagementUserJWT(h.tokenSignKey), h.getRenewToken)
}

type SignInRequest struct {
	Sub        string   `json:"sub"`
	Roles      []string `json:"roles"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	RenewToken string   `json:"renewToken"`
	InstanceID string   `json:"instanceId"`
}

func (h *HttpEndpoints) signInWithIdP(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("signInWithIdP: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slog.Info("signInWithIdP called with ", slog.String("sub", req.Sub))

	for _, role := range req.Roles {
		slog.Info("Role: ", slog.String("role", role))
	}

	// Use req to access the request body data
	// ...

	c.JSON(http.StatusNotImplemented, gin.H{"error": "unimplemented"})
}

func (h *HttpEndpoints) getRenewToken(c *gin.Context) {
	// TODO: get user id from jwt
	// TODO: look up if user has a valid renew token
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	slog.Info("getRenewToken called with ", slog.String("id", token.ID))

	c.JSON(http.StatusNotImplemented, gin.H{"error": "unimplemented"})
}
