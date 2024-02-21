package apihandlers

import (
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	mUserDB "github.com/case-framework/case-backend/pkg/db/management-user"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddManagementAuthAPI(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/signin-with-idp", mw.RequirePayload(), h.signInWithIdP)
	auth.GET("/renew-token/:sessionID",
		mw.GetAndValidateManagementUserJWT(h.tokenSignKey),
		mw.IsInstanceIDInJWTAllowed(h.allowedInstanceIDs),
		h.getRenewToken,
	)
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

	if !h.isInstanceAllowed(req.InstanceID) {
		slog.Warn("signInWithIdP: instance not allowed", slog.String("instanceID", req.InstanceID))
		c.JSON(http.StatusForbidden, gin.H{"error": "instance not allowed"})
		return
	}

	if req.Sub == "" {
		slog.Warn("signInWithIdP: no sub")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing sub"})
		return
	}

	isAdmin := false
	for _, role := range req.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	// Find user in database
	existingUser, err := h.muDBConn.GetUserBySub(req.InstanceID, req.Sub)
	if err != nil || existingUser == nil {
		slog.Info("sign up with a new management user", slog.String("sub", req.Sub), slog.String("instanceID", req.InstanceID), slog.String("name", req.Name), slog.String("email", req.Email))
		// Create new user
		existingUser, err = h.muDBConn.CreateUser(req.InstanceID, &mUserDB.ManagementUser{
			Sub:         req.Sub,
			Username:    req.Name,
			Email:       req.Email,
			IsAdmin:     isAdmin,
			LastLoginAt: time.Now(),
		})
		if err != nil {
			slog.Error("could not create new user", slog.String("sub", req.Sub), slog.String("instanceID", req.InstanceID), slog.String("name", req.Name), slog.String("email", req.Email), slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create new user"})
			return
		}
	} else {
		slog.Info("sign in with an existing management user", slog.String("sub", req.Sub), slog.String("instanceID", req.InstanceID), slog.String("name", req.Name), slog.String("email", req.Email))
		// Update existing user
		err = h.muDBConn.UpdateUser(req.InstanceID, existingUser.ID.Hex(), req.Email, req.Name, isAdmin, time.Now())
		if err != nil {
			slog.Error("could not update existing user", slog.String("sub", req.Sub), slog.String("instanceID", req.InstanceID), slog.String("name", req.Name), slog.String("email", req.Email), slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update existing user"})
			return
		}
	}

	sessionId := ""

	// Create new session
	if req.RenewToken != "" {
		session, err := h.muDBConn.CreateSession(req.InstanceID, existingUser.ID.Hex(), req.RenewToken)
		if err != nil {
			slog.Error("could not create session", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
			return
		}
		sessionId = session.ID.Hex()
	}

	// generate new JWT token
	token, err := jwthandling.GenerateNewManagementUserToken(
		h.tokenExpiresIn,
		existingUser.ID.Hex(),
		req.InstanceID,
		isAdmin,
		map[string]string{},
		h.tokenSignKey,
	)
	if err != nil {
		slog.Error("signInWithIdP: could not generate token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"sessionID":   sessionId,
		"expiresAt":   time.Now().Add(h.tokenExpiresIn).Unix(),
		"isAdmin":     isAdmin,
	})
}

// getRenewToken to get a the renew token for the user
func (h *HttpEndpoints) getRenewToken(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		slog.Warn("getRenewToken: no sessionID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no sessionID"})
		return
	}

	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	existingSession, err := h.muDBConn.GetSession(token.InstanceID, sessionID)
	if err != nil {
		slog.Error("getRenewToken: could not get session", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get session"})
		return
	}
	if existingSession.UserID != token.ID {
		slog.Warn("getRenewToken: user not allowed to get renew token", slog.String("userID", token.ID), slog.String("sessionUserID", existingSession.UserID))
		c.JSON(http.StatusForbidden, gin.H{"error": "user not allowed to get renew token"})
		return
	}

	// delete the session
	err = h.muDBConn.DeleteSession(token.InstanceID, sessionID)
	if err != nil {
		slog.Error("getRenewToken: could not delete session", slog.String("error", err.Error()))
	}

	c.JSON(http.StatusOK, gin.H{"renewToken": existingSession.RenewToken})
}
