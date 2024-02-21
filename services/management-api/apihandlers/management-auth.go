package apihandlers

import (
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddManagementAuthAPI(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/signin-with-idp", mw.RequirePayload(), h.signInWithIdP)
	auth.GET("/renew-token", mw.GetAndValidateManagementUserJWT(h.tokenSignKey), h.getRenewToken)
}

// SignInRequest is the request body for the signin-with-idp endpoint
type SignInRequest struct {
	Sub        string   `json:"sub"`
	Roles      []string `json:"roles"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	RenewToken string   `json:"renewToken"`
	InstanceID string   `json:"instanceId"`
}

// singInWithIdP to generate a new token and update the user in the database
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

	// TODO: look up user in database by instanceID and sub
	// TODO: if user exists, update user with new token, email, and name
	// TODO: if user does not exist, create new user with token, email, name

	// TODO: generate new JWT token
	token, err := jwthandling.GenerateNewManagementUserToken(
		5*time.Minute,
		"testUserID",
		"testInstanceID",
		false,
		map[string]string{},
		h.tokenSignKey,
	)
	if err != nil {
		slog.Error("signInWithIdP: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	slog.Info("signInWithIdP: generated token ", slog.String("token", token))

	// Use req to access the request body data
	// ...

	c.JSON(http.StatusNotImplemented, gin.H{"error": "unimplemented"})
}

// getRenewToken to get a the renew token for the user
func (h *HttpEndpoints) getRenewToken(c *gin.Context) {
	// TODO: get user id from jwt
	// TODO: look up if user has a valid renew token
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	slog.Info("getRenewToken called with ", slog.String("id", token.ID))

	c.JSON(http.StatusNotImplemented, gin.H{"error": "unimplemented"})
}
