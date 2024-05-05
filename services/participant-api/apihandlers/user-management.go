package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

const (
	MAX_PROFILES_ALLOWED = 6
)

func (h *HttpEndpoints) AddUserManagementAPI(rg *gin.RouterGroup) {
	userGroup := rg.Group("/user")
	userGroup.Use(mw.GetAndValidateParticipantUserJWT(h.tokenSignKey))
	{
		userGroup.GET("/", h.getUser)
		userGroup.POST("/profiles", mw.RequirePayload(), h.addNewProfileHandl)
		userGroup.PUT("/profiles", mw.RequirePayload(), h.updateProfileHandl)
		//userGroup.POST("/profile/remove", mw.RequirePayload(), h.removeProfileHandl)
	}
}

func (h *HttpEndpoints) getUser(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get user"})
		return
	}
	user.Account.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *HttpEndpoints) addNewProfileHandl(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var profile userTypes.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind profile"})
		return
	}

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	if len(user.Profiles) > MAX_PROFILES_ALLOWED {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reached profile limit"})
		return
	}
	user.AddProfile(profile)

	_, err = h.userDBConn.ReplaceUser(token.InstanceID, user)
	if err != nil {
		slog.Error("cannot update user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile})
}

func (h *HttpEndpoints) updateProfileHandl(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var profile userTypes.Profile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind profile"})
		return
	}

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	err = user.UpdateProfile(profile)
	if err != nil {
		slog.Error("cannot update profile", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update profile"})
		return
	}

	_, err = h.userDBConn.ReplaceUser(token.InstanceID, user)
	if err != nil {
		slog.Error("cannot update user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile})
}
