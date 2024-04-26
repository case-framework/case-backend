package apihandlers

import (
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/case-framework/case-backend/pkg/user-management/pwhash"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

const (
	loginFailedAttemptWindow = 5 * 60 // to count the login failures, seconds
	allowedPasswordAttempts  = 10

	signupRateLimitWindow = 5 * 60 // to count the new signups, seconds
)

func (h *HttpEndpoints) AddParticipantAuthAPI(rg *gin.RouterGroup) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/login", mw.RequirePayload(), h.loginWithEmail)
		authGroup.POST("/signup", mw.RequirePayload(), h.signupWithEmail)
		authGroup.POST("/token/renew", mw.RequirePayload(), mw.GetAndValidateParticipantUserJWT(h.tokenSignKey), h.refreshToken)
		authGroup.GET("/token/validate", mw.RequirePayload(), mw.GetAndValidateParticipantUserJWT(h.tokenSignKey), h.validateToken)
		authGroup.GET("/token/revoke", mw.GetAndValidateParticipantUserJWT(h.tokenSignKey), h.revokeRefreshTokens)
	}

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

	if req.Email == "" || req.Password == "" || req.InstanceID == "" {
		slog.Error("missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	if !h.isInstanceAllowed(req.InstanceID) {
		slog.Error("instance not allowed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid instance id"})
		return
	}

	req.Email = umUtils.SanitizeEmail(req.Email)

	user, err := h.userDBConn.GetUserByAccountID(req.InstanceID, req.Email)
	if err != nil {
		slog.Warn("login attempt with wrong email address", slog.String("email", req.Email), slog.String("instanceID", req.InstanceID), slog.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if umUtils.HasMoreAttemptsRecently(user.Account.FailedLoginAttempts, allowedPasswordAttempts, loginFailedAttemptWindow) {
		slog.Warn("login attempt with too many failed attempts", slog.String("email", req.Email), slog.String("instanceID", req.InstanceID))

		if err := h.userDBConn.SaveFailedLoginAttempt(req.InstanceID, user.ID.Hex()); err != nil {
			slog.Error("failed to save failed login attempt", slog.String("error", err.Error()))
		}
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	match, err := pwhash.ComparePasswordWithHash(user.Account.Password, req.Password)
	if err != nil || !match {
		if err == nil {
			err = errors.New("passwords do not match")
		}
		slog.Warn("login attempt with wrong password", slog.String("email", req.Email), slog.String("instanceID", req.InstanceID), slog.String("error", err.Error()))
		if err := h.userDBConn.SaveFailedLoginAttempt(req.InstanceID, user.ID.Hex()); err != nil {
			slog.Error("failed to save failed login attempt", slog.String("error", err.Error()))
		}
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// generate jwt
	mainProfileID, otherProfileIDs := umUtils.GetMainAndOtherProfiles(user)

	token, err := jwthandling.GenerateNewParticipantUserToken(
		h.ttls.AccessToken,
		user.ID.Hex(),
		req.InstanceID,
		mainProfileID,
		map[string]string{},
		user.Account.AccountConfirmedAt > 0,
		nil,
		otherProfileIDs,
		h.tokenSignKey,
		nil,
	)
	if err != nil {
		slog.Error("failed to generate token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// generate refresh token
	renewToken, err := umUtils.GenerateUniqueTokenString()
	if err != nil {
		slog.Error("failed to generate renew token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = h.userDBConn.CreateRenewToken(req.InstanceID, user.ID.Hex(), renewToken, 0)
	if err != nil {
		slog.Error("failed to save renew token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// update timestamps
	user.Timestamps.LastLogin = time.Now().Unix()
	user.Timestamps.MarkedForDeletion = 0
	user.Account.VerificationCode = userTypes.VerificationCode{}
	user.Account.FailedLoginAttempts = umUtils.RemoveAttemptsOlderThan(user.Account.FailedLoginAttempts, 3600)
	user.Account.PasswordResetTriggers = umUtils.RemoveAttemptsOlderThan(user.Account.PasswordResetTriggers, 7200)

	user, err = h.userDBConn.ReplaceUser(req.InstanceID, user)
	if err != nil {
		slog.Error("failed to update user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// cleanup tokens for password reset (user can login now...)
	if err := h.globalInfosDBConn.DeleteAllTempTokenForUser(
		req.InstanceID,
		user.ID.Hex(),
		userTypes.TOKEN_PURPOSE_PASSWORD_RESET,
	); err != nil {
		slog.Error("failed to delete temp tokens", slog.String("error", err.Error()))
	}

	slog.Info("login successful", slog.String("subject", user.ID.Hex()), slog.String("instanceID", req.InstanceID))

	user.Account.Password = ""
	user.Account.VerificationCode = userTypes.VerificationCode{}

	c.JSON(http.StatusOK, gin.H{
		"token": gin.H{
			"accessToken":     token,
			"refreshToken":    renewToken,
			"expiresIn":       h.ttls.AccessToken.Seconds(),
			"selectedProfile": mainProfileID,
		},
		"user": user,
	})
}

type SignupWithEmailReq struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	InstanceID        string `json:"instanceId"`
	InfoCheck         string `json:"infoCheck"`
	PreferredLanguage string `json:"preferredLanguage"`
}

func (h *HttpEndpoints) signupWithEmail(c *gin.Context) {
	var req SignupWithEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email == "" || req.Password == "" || req.InstanceID == "" {
		slog.Error("missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	if req.InfoCheck != "" {
		slog.Warn("honeypot field filled out", slog.String("email", req.Email), slog.String("instanceID", req.InstanceID), slog.String("infoCheck", req.InfoCheck))
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid request"})
		return
	}

	if !h.isInstanceAllowed(req.InstanceID) {
		slog.Error("instance not allowed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid instance id"})
		return
	}

	req.Email = umUtils.SanitizeEmail(req.Email)

	if !umUtils.CheckEmailFormat(req.Email) {
		slog.Error("invalid email format", slog.String("email", req.Email))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	if !umUtils.CheckPasswordFormat(req.Password) {
		slog.Error("invalid password format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password format"})
		return
	}

	if !umUtils.CheckLanguageCode(req.PreferredLanguage) {
		slog.Error("invalid preferred language code", slog.String("preferredLanguage", req.PreferredLanguage))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid preferred language code"})
		return
	}

	// rate limit
	newUserCount, err := h.userDBConn.CountRecentlyCreatedUsers(req.InstanceID, signupRateLimitWindow)
	if err != nil {
		slog.Error("failed to count new users", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if newUserCount >= int64(h.maxNewUsersPer5Minute) {
		slog.Warn("rate limit for new users reached", slog.String("instanceID", req.InstanceID))
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "try again later"})
		return
	}

	// hash password
	password, err := pwhash.HashPassword(req.Password)
	if err != nil {
		slog.Error("failed to hash password", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// create user
	newUser := umUtils.InitNewEmailUser(req.Email, password, req.PreferredLanguage)
	id, err := h.userDBConn.AddUser(req.InstanceID, newUser)
	if err != nil {
		slog.Error("failed to create new user", slog.String("error", err.Error()))
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return

	}
	newUser.ID, _ = primitive.ObjectIDFromHex(id)

	// contact verification in go routine
	go h.prepAndSendEmailVerification(
		newUser.ID.Hex(),
		req.InstanceID,
		req.Email,
		req.PreferredLanguage,
		h.ttls.EmailContactVerificationToken,
	)

	// generate jwt
	mainProfileID, otherProfileIDs := umUtils.GetMainAndOtherProfiles(newUser)

	token, err := jwthandling.GenerateNewParticipantUserToken(
		h.ttls.AccessToken,
		newUser.ID.Hex(),
		req.InstanceID,
		mainProfileID,
		map[string]string{},
		newUser.Account.AccountConfirmedAt > 0,
		nil,
		otherProfileIDs,
		h.tokenSignKey,
		nil,
	)
	if err != nil {
		slog.Error("failed to generate token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// generate refresh token
	renewToken, err := umUtils.GenerateUniqueTokenString()
	if err != nil {
		slog.Error("failed to generate renew token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// generate refresh token
	err = h.userDBConn.CreateRenewToken(req.InstanceID, newUser.ID.Hex(), renewToken, 0)
	if err != nil {
		slog.Error("failed to save renew token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// return tokens and user
	slog.Info("login successful", slog.String("subject", newUser.ID.Hex()), slog.String("instanceID", req.InstanceID))

	newUser.Account.Password = ""
	newUser.Account.VerificationCode = userTypes.VerificationCode{}

	c.JSON(http.StatusOK, gin.H{
		"token": gin.H{
			"accessToken":     token,
			"refreshToken":    renewToken,
			"expiresIn":       h.ttls.AccessToken.Seconds(),
			"selectedProfile": mainProfileID,
		},
		"user": newUser,
	})
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *HttpEndpoints) refreshToken(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var req RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if user still exists
	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Warn("user not found", slog.String("subject", token.Subject), slog.String("instanceID", token.InstanceID), slog.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// generate new refresh token
	newRenewToken, err := umUtils.GenerateUniqueTokenString()
	if err != nil {
		slog.Error("failed to generate renew token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Check if previous token is still valid
	rt, err := h.userDBConn.FindAndUpdateRenewToken(
		token.InstanceID,
		token.Subject,
		req.RefreshToken,
		newRenewToken,
	)
	if err != nil {
		slog.Error("failed to find and update renew token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if rt.NextToken == newRenewToken {
		// this is the first time the refresh token is used
		err := h.userDBConn.CreateRenewToken(token.InstanceID, token.Subject, newRenewToken, 0)
		if err != nil {
			slog.Error("failed to save renew token", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	} else {
		newRenewToken = rt.NextToken
	}

	// update timestamps (last token refresh, reset markeed for deletion, etc.)
	err = h.userDBConn.UpdateUser(token.InstanceID, token.Subject, bson.M{
		"$set": bson.M{
			"timestamps.lastTokenRefresh":  time.Now().Unix(),
			"timestamps.markedForDeletion": 0,
		},
	})
	if err != nil {
		slog.Error("failed to update user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// generate jwt
	mainProfileID, otherProfileIDs := umUtils.GetMainAndOtherProfiles(user)

	newJwt, err := jwthandling.GenerateNewParticipantUserToken(
		h.ttls.AccessToken,
		user.ID.Hex(),
		token.InstanceID,
		mainProfileID,
		map[string]string{},
		user.Account.AccountConfirmedAt > 0,
		nil,
		otherProfileIDs,
		h.tokenSignKey,
		token.LastOTPProvided,
	)
	if err != nil {
		slog.Error("failed to generate token", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	user.Account.Password = ""
	user.Account.VerificationCode = userTypes.VerificationCode{}

	c.JSON(http.StatusOK, gin.H{
		"token": gin.H{
			"accessToken":     newJwt,
			"refreshToken":    newRenewToken,
			"expiresIn":       h.ttls.AccessToken.Seconds(),
			"selectedProfile": mainProfileID,
		},
		"user": user,
	})
}

func (h *HttpEndpoints) validateToken(c *gin.Context) {
	// read validated token
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	// check if user still exists
	_, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Warn("user not found", slog.String("subject", token.Subject), slog.String("instanceID", token.InstanceID), slog.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokenInfos": token})
}

func (h *HttpEndpoints) revokeRefreshTokens(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	count, err := h.userDBConn.DeleteRenewTokensForUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("failed to delete renew tokens", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	slog.Debug("deleted renew tokens", slog.Int64("count", count))
	c.JSON(http.StatusOK, gin.H{"message": "tokens revoked"})
}
