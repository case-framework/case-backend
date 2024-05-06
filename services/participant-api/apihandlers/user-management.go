package apihandlers

import (
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	emailTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	studyService "github.com/case-framework/case-backend/pkg/study"
	"github.com/case-framework/case-backend/pkg/user-management/pwhash"
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
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
		userGroup.POST("/profiles/remove", mw.RequirePayload(), h.removeProfileHandl)

		userGroup.POST("/password", mw.RequirePayload(), h.changePasswordHandl)
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

func (h *HttpEndpoints) removeProfileHandl(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind profile"})
		return
	}

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	err = user.RemoveProfile(req.ProfileID)
	if err != nil {
		slog.Error("cannot remove profile", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot remove profile"})
		return
	}

	_, err = h.userDBConn.ReplaceUser(token.InstanceID, user)
	if err != nil {
		slog.Error("cannot update user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update user"})
		return
	}

	studyService.OnProfileDeleted(token.InstanceID, req.ProfileID)

	c.JSON(http.StatusOK, gin.H{"message": "profile removed"})
}

func (h *HttpEndpoints) changePasswordHandl(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind profile"})
		return
	}

	// check password format
	if !umUtils.CheckPasswordFormat(req.NewPassword) {
		slog.Error("invalid password format", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", "invalid password format"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password format"})
		return
	}

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	match, err := pwhash.ComparePasswordWithHash(user.Account.Password, req.OldPassword)
	if err != nil || !match {
		slog.Error("old password does not match", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	hashedPassword, err := pwhash.HashPassword(req.NewPassword)
	if err != nil {
		slog.Error("cannot hash password", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot hash password"})
		return
	}

	update := bson.M{"$set": bson.M{"account.password": hashedPassword, "timestamps.lastPasswordChange": time.Now().Unix()}}
	if err := h.userDBConn.UpdateUser(token.InstanceID, user.ID.Hex(), update); err != nil {
		slog.Error("cannot update user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update user"})
		return
	}

	go h.sendSimpleEmail(
		token.InstanceID,
		[]string{user.Account.AccountID},
		emailTypes.EMAIL_TYPE_PASSWORD_CHANGED,
		"",
		user.Account.PreferredLanguage,
		nil,
		true,
	)

	slog.Info("password changed successful", slog.String("userID", user.ID.Hex()), slog.String("instanceID", token.InstanceID))

	if err := h.globalInfosDBConn.DeleteAllTempTokenForUser(token.InstanceID, user.ID.Hex(), userTypes.TOKEN_PURPOSE_PASSWORD_RESET); err != nil {
		slog.Error("failed to delete temp tokens", slog.String("error", err.Error()))
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}
