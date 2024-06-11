package apihandlers

import (
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	emailTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
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

		userGroup.POST("/change-account-email", mw.RequirePayload(), h.changeAccountEmailHandl)

		userGroup.DELETE("/", h.deleteUser)
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
		ProfileID          string                     `json:"profileId"`
		ExitSurveyResponse *studyTypes.SurveyResponse `json:"exitSurveyResponse"`
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

	studyService.OnProfileDeleted(token.InstanceID, req.ProfileID, req.ExitSurveyResponse)

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

	if umUtils.IsPasswordOnBlocklist(req.NewPassword) {
		slog.Error("password on blocklist", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", "password on blocklist"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "password on blocklist"})
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

func (h *HttpEndpoints) changeAccountEmailHandl(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind profile"})
		return
	}

	req.Email = umUtils.SanitizeEmail(req.Email)

	if !umUtils.CheckEmailFormat(req.Email) {
		slog.Error("invalid email format", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", "invalid email format"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	if user.Account.AccountID == req.Email {
		slog.Error("cannot change account email to self", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change account email to self"})
		return
	}

	match, err := pwhash.ComparePasswordWithHash(user.Account.Password, req.Password)
	if err != nil || !match {
		slog.Error("password does not match", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	// is email already in use?
	_, err = h.userDBConn.GetUserByAccountID(token.InstanceID, req.Email)
	if err == nil {
		slog.Error("email already in use", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("email", req.Email))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	if user.Account.Type != userTypes.ACCOUNT_TYPE_EMAIL {
		slog.Error("account type not email", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", "account type not email"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "account type not email"})
		return
	}

	oldCI, oldFound := user.FindContactInfoByTypeAndAddr("email", user.Account.AccountID)
	if !oldFound {
		slog.Error("old contact info not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", "old contact info not found"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "old contact info not found"})
		return
	}

	if user.Account.AccountConfirmedAt > 0 {
		// old account is confirmed already
		go h.prepTokenAndSendEmail(
			user.ID.Hex(),
			token.InstanceID,
			oldCI.Email,
			user.Account.PreferredLanguage,
			userTypes.TOKEN_PURPOSE_RESTORE_ACCOUNT_ID,
			h.ttls.EmailContactVerificationToken,
			emailTypes.EMAIL_TYPE_ACCOUNT_ID_CHANGED,
			map[string]string{
				"newEmail": req.Email,
			},
		)
	}

	// update user
	if user.Profiles[0].Alias == umUtils.BlurEmailAddress(user.Account.AccountID) {
		user.Profiles[0].Alias = umUtils.BlurEmailAddress(req.Email)
	}
	user.Account.AccountID = req.Email
	user.Account.AccountConfirmedAt = -1

	// Add new address to contact list if necessary:
	ci, found := user.FindContactInfoByTypeAndAddr("email", req.Email)
	if found {
		// new email already confirmed
		if ci.ConfirmedAt > 0 {
			user.Account.AccountConfirmedAt = ci.ConfirmedAt
		}
	} else {
		user.AddNewEmail(req.Email, false)
	}

	newCI, newFound := user.FindContactInfoByTypeAndAddr("email", req.Email)
	if !newFound {
		slog.Error("new contact info not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", "new contact info not found"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "new contact info not found"})
		return
	}

	user.ReplaceContactInfoInContactPreferences(oldCI.ID.Hex(), newCI.ID.Hex())

	// start confirmation workflow of necessary:
	if user.Account.AccountConfirmedAt <= 0 {
		go h.prepTokenAndSendEmail(
			user.ID.Hex(),
			token.InstanceID,
			user.Account.AccountID,
			user.Account.PreferredLanguage,
			userTypes.TOKEN_PURPOSE_CONTACT_VERIFICATION,
			h.ttls.EmailContactVerificationToken,
			emailTypes.EMAIL_TYPE_VERIFY_EMAIL,
			nil,
		)
	}

	err = user.RemoveContactInfo(oldCI.ID.Hex())
	if err != nil {
		slog.Error("cannot remove old contact info", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
	}

	_, err = h.userDBConn.ReplaceUser(token.InstanceID, user)
	if err != nil {
		slog.Error("cannot update user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account email changed"})
}

func (h *HttpEndpoints) deleteUser(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var req struct {
		ExitSurveyResponse *studyTypes.SurveyResponse `json:"exitSurveyResponse"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	for _, profile := range user.Profiles {
		var exitResp *studyTypes.SurveyResponse
		if profile.MainProfile {
			exitResp = req.ExitSurveyResponse
		} else {
			exitResp = nil
		}
		studyService.OnProfileDeleted(token.InstanceID, profile.ID.Hex(), exitResp)
	}

	// delete all temp tokens
	err = h.globalInfosDBConn.DeleteAllTempTokenForUser(token.InstanceID, user.ID.Hex(), "")
	if err != nil {
		slog.Error("failed to delete temp tokens", slog.String("error", err.Error()))
	}

	h.sendSimpleEmail(
		token.InstanceID,
		[]string{user.Account.AccountID},
		emailTypes.EMAIL_TYPE_ACCOUNT_DELETED,
		"",
		user.Account.PreferredLanguage,
		nil,
		true,
	)

	err = h.userDBConn.DeleteUser(token.InstanceID, user.ID.Hex())
	if err != nil {
		slog.Error("cannot delete user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete user"})
		return
	}

	slog.Info("user deleted successful", slog.String("userID", user.ID.Hex()), slog.String("instanceID", token.InstanceID))

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
