package apihandlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/case-framework/case-backend/pkg/messaging/sms"
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
	MAX_PROFILES_ALLOWED                          = 6
	MAX_PHONE_NUMBER_VERIFICATION_REQUEST_PER_24H = 10
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
		userGroup.POST("/change-phone-number", mw.RequirePayload(), h.updatePhoneNumberHandler)
		userGroup.GET("/request-phone-number-verification", h.requestPhoneNumberVerificationHandl)
		userGroup.POST("/verify-phone-number", mw.RequirePayload(), h.verifyPhoneNumberHandler)

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

func (h *HttpEndpoints) updatePhoneNumberHandler(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var req struct {
		NewPhoneNumber string `json:"newPhoneNumber"`
		Password       string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind profile"})
		return
	}

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		randomWait(5, 10)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	match, err := pwhash.ComparePasswordWithHash(user.Account.Password, req.Password)
	if err != nil || !match {
		slog.Error("password does not match", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		randomWait(5, 10)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	// if have too many phone numbers within the last 24 hours, return error
	count, err := h.messagingDBConn.CountSentSMSForUser(token.InstanceID, token.Subject, sms.SMS_MESSAGE_TYPE_VERIFY_PHONE_NUMBER, time.Now().Add(-time.Hour*24))
	if err != nil {
		slog.Error("failed to count sent SMS", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
	}
	if count > MAX_PHONE_NUMBER_VERIFICATION_REQUEST_PER_24H || err != nil {
		slog.Warn("too many phone numbers sent within the last 24 hours", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		randomWait(5, 10)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many phone numbers sent within the last 24 hours"})
		return
	}

	// check if phone number is already set
	phoneNumber := umUtils.SanitizePhoneNumber(req.NewPhoneNumber)
	currentPhoneNumber, err := user.GetPhoneNumber()
	if err == nil && currentPhoneNumber.Phone == phoneNumber {
		slog.Error("phone number is already set", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		randomWait(5, 10)
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone number is already set"})
		return
	}

	user.SetPhoneNumber(req.NewPhoneNumber)

	// send email to user about phone number change
	if user.Account.AccountConfirmedAt > 0 {
		// old account is confirmed already
		go h.prepTokenAndSendEmail(
			user.ID.Hex(),
			token.InstanceID,
			user.Account.AccountID,
			user.Account.PreferredLanguage,
			userTypes.TOKEN_PURPOSE_RESTORE_ACCOUNT_ID,
			h.ttls.EmailContactVerificationToken,
			emailTypes.EMAIL_TYPE_PHONE_NUMBER_CHANGED,
			map[string]string{
				"newPhoneNumber": req.NewPhoneNumber,
			},
		)
	}

	_, err = h.userDBConn.ReplaceUser(token.InstanceID, user)
	if err != nil {
		slog.Error("cannot update user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update user"})
		return
	}
	slog.Info("phone number changed", slog.String("instanceId", token.InstanceID), slog.String("userID", token.Subject))

	c.JSON(http.StatusOK, gin.H{"message": "phone number changed"})
}

func (h *HttpEndpoints) requestPhoneNumberVerificationHandl(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		slog.Error("user not found", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		randomWait(5, 10)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	if user.Account.AccountConfirmedAt < 1 {
		slog.Error("account not confirmed", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		randomWait(5, 10)
		c.JSON(http.StatusBadRequest, gin.H{"error": "account not confirmed"})
		return
	}

	// check daily limit
	count24h, err := h.messagingDBConn.CountSentSMSForUser(token.InstanceID, token.Subject, sms.SMS_MESSAGE_TYPE_VERIFY_PHONE_NUMBER, time.Now().Add(-time.Hour*24))
	if err != nil {
		slog.Error("failed to count sent SMS", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
	}
	if count24h > MAX_PHONE_NUMBER_VERIFICATION_REQUEST_PER_24H || err != nil {
		slog.Warn("too many phone numbers sent within the last 24 hours", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		randomWait(5, 10)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many phone numbers sent within the last 24 hours"})
		return
	}

	// check short term limit
	countShortTerm, err := h.messagingDBConn.CountSentSMSForUser(token.InstanceID, token.Subject, sms.SMS_MESSAGE_TYPE_VERIFY_PHONE_NUMBER, time.Now().Add(-time.Second*15))
	if err != nil {
		slog.Error("failed to count sent SMS", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
	}
	if countShortTerm > 0 || err != nil {
		slog.Warn("already sent an SMS within the last 15 seconds", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		randomWait(5, 10)
		c.JSON(http.StatusOK, gin.H{"message": "already sent an SMS within the last 15 seconds"})
		return
	}

	// check if phone number is already verified
	phoneContact, err := user.GetPhoneNumber()
	if err != nil {
		slog.Error("failed to get phone number, maybe it is not set", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		randomWait(5, 10)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get phone number"})
		return
	}

	if phoneContact.ConfirmedAt > 0 {
		slog.Error("phone number already verified", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject))
		randomWait(5, 10)
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone number already verified"})
		return
	}

	// generate OTP
	code, err := umUtils.GenerateOTPCode(6)
	if err != nil {
		slog.Error("failed to generate OTP", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate OTP"})
		return
	}

	// save OTP
	err = h.userDBConn.CreateOTP(token.InstanceID, token.Subject, code, userTypes.SMSOTP)
	if err != nil {
		slog.Error("failed to save OTP", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save OTP"})
		return
	}

	half := len(code) / 2
	formattedCode := fmt.Sprintf("%s-%s", code[:half], code[half:])

	// send SMS
	err = sms.SendSMS(token.InstanceID, phoneContact.Phone, token.Subject, sms.SMS_MESSAGE_TYPE_VERIFY_PHONE_NUMBER, user.Account.PreferredLanguage, map[string]string{
		"verificationCode": formattedCode,
	})
	if err != nil {
		slog.Error("failed to send SMS", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send SMS"})
		return
	}
	slog.Info("sent SMS for phone number verification", slog.String("instanceId", token.InstanceID), slog.String("userID", token.Subject))
	c.JSON(http.StatusOK, gin.H{"message": "SMS sent"})
}

func (h *HttpEndpoints) verifyPhoneNumberHandler(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind profile"})
		return
	}

	slog.Info("verifying account phone number", slog.String("instanceId", token.InstanceID), slog.String("userID", token.Subject))
	// TODO: lookup user
	// TODO: lookup code
	// TODO: check if code is valid
	// TODO: check if phone is not already verified
	// TODO: update phone

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
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
