package apihandlers

import (
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
	"github.com/gin-gonic/gin"

	emailTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

const (
	passwordResetAttemptWindow   = 60 * 60 // seconds
	PASSWWORD_RESET_MAX_ATTEMPTS = 5

	PASSWORD_RESET_TOKEN_TTL = 24 * time.Hour
)

func (h *HttpEndpoints) AddPasswordResetAPI(rg *gin.RouterGroup) {
	pwResetGroup := rg.Group("/password-reset")
	{
		pwResetGroup.POST("/initiate", mw.RequirePayload(), h.initiatePasswordReset)
		//pwResetGroup.POST("/get-infos", mw.RequirePayload(), h.getPasswordResetInfos)
		// pwResetGroup.POST("/reset", mw.RequirePayload(), h.resetPassword)
	}
}

func (h *HttpEndpoints) initiatePasswordReset(c *gin.Context) {
	var req struct {
		Email      string `json:"email"`
		InstanceID string `json:"instanceID"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("bad request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.isInstanceAllowed(req.InstanceID) {
		slog.Error("instance not allowed", slog.String("instanceID", req.InstanceID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid instance id"})
		return
	}

	req.Email = umUtils.SanitizeEmail(req.Email)

	user, err := h.userDBConn.GetUserByAccountID(req.InstanceID, req.Email)
	if err != nil {
		slog.Warn("password reset for non-existing user", slog.String("email", req.Email), slog.String("instanceID", req.InstanceID), slog.String("error", err.Error()))
		randomWait(10)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if umUtils.HasMoreAttemptsRecently(user.Account.PasswordResetTriggers, PASSWWORD_RESET_MAX_ATTEMPTS, passwordResetAttemptWindow) {
		slog.Warn("password reset rate limited", slog.String("email", req.Email), slog.String("instanceID", req.InstanceID))
		randomWait(10)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limited"})
		return
	}

	go h.prepTokenAndSendEmail(
		user.ID.Hex(),
		req.InstanceID,
		user.Account.AccountID,
		user.Account.PreferredLanguage,
		userTypes.TOKEN_PURPOSE_PASSWORD_RESET,
		PASSWORD_RESET_TOKEN_TTL,
		emailTypes.EMAIL_TYPE_PASSWORD_RESET,
		map[string]string{
			"validUntil": "24",
		},
	)

	if err := h.userDBConn.SavePasswordResetTrigger(
		req.InstanceID,
		user.ID.Hex(),
	); err != nil {
		slog.Error("failed to save password reset trigger", slog.String("error", err.Error()))
	}

	slog.Info("password reset initiated", slog.String("email", req.Email), slog.String("instanceID", req.InstanceID))
	c.JSON(http.StatusOK, gin.H{"message": "password reset initiated"})
}
