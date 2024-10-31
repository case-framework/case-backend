package usermanagement

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	"github.com/case-framework/case-backend/pkg/messaging/sms"
	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	"github.com/case-framework/case-backend/pkg/user-management/utils"
)

const (
	MAX_OTP_ATTEMPTS = 10
	OTP_LENGTH       = 6
)

var (
	pUserDBService        *userDB.ParticipantUserDBService
	globalInfosDBServices *globalinfosDB.GlobalInfosDBService
)

func Init(
	participantUserDBService *userDB.ParticipantUserDBService,
	globalInfosDBService *globalinfosDB.GlobalInfosDBService,
) {
	pUserDBService = participantUserDBService
	globalInfosDBServices = globalInfosDBService
}

func SendOTPByEmail(
	instanceID,
	userID string,
	sendEmail func(email string, code string, preferredLang string, expiresAt int64) error,
) error {
	// check count of recent attempts
	count, err := pUserDBService.CountOTP(instanceID, userID)
	if err != nil {
		return err
	}

	if count >= MAX_OTP_ATTEMPTS {
		slog.Warn("too many OTP requests", slog.String("instanceID", instanceID), slog.String("userID", userID))
		return errors.New("too many attempts")
	}

	otp, err := pUserDBService.GetLastOTP(instanceID, userID, string(userTypes.EmailOTP))
	if err == nil && otp.CreatedAt.After(time.Now().Add(-time.Second*30)) {
		// last OTP was sent less than 30 seconds ago, so don't send another one - for rate limiting
		slog.Debug("last OTP was sent less than 30 seconds ago", slog.String("instanceID", instanceID), slog.String("userID", userID))
		return nil
	}

	user, err := pUserDBService.GetUser(instanceID, userID)
	if err != nil {
		slog.Error("error getting user", slog.String("instanceID", instanceID), slog.String("userID", userID), slog.String("error", err.Error()))
		return err
	}

	// generate OTP
	code, err := utils.GenerateOTPCode(OTP_LENGTH)
	if err != nil {
		return err
	}

	// save OTP
	err = pUserDBService.CreateOTP(instanceID, userID, code, userTypes.EmailOTP, MAX_OTP_ATTEMPTS)
	if err != nil {
		return err
	}

	half := len(code) / 2
	formattedCode := fmt.Sprintf("%s-%s", code[:half], code[half:])

	// send OTP
	err = sendEmail(user.Account.AccountID, formattedCode, user.Account.PreferredLanguage, time.Now().Add(time.Second*userDB.OTP_TTL).Unix())
	if err != nil {
		return err
	}

	return nil
}

func SendOTPBySMS(instanceID, userID string) error {
	// check count of recent attempts
	count, err := pUserDBService.CountOTP(instanceID, userID)
	if err != nil {
		return err
	}

	if count >= MAX_OTP_ATTEMPTS {
		slog.Warn("too many OTP requests", slog.String("instanceID", instanceID), slog.String("userID", userID))
		return errors.New("too many attempts")
	}

	otp, err := pUserDBService.GetLastOTP(instanceID, userID, string(userTypes.SMSOTP))
	if err == nil && otp.CreatedAt.After(time.Now().Add(-time.Second*30)) {
		// last OTP was sent less than 30 seconds ago, so don't send another one - for rate limiting
		slog.Debug("last OTP was sent less than 30 seconds ago", slog.String("instanceID", instanceID), slog.String("userID", userID))
		return nil
	}

	user, err := pUserDBService.GetUser(instanceID, userID)
	if err != nil {
		slog.Error("error getting user", slog.String("instanceID", instanceID), slog.String("userID", userID), slog.String("error", err.Error()))
		return err
	}

	phoneInfos, err := user.GetPhoneNumber()
	if err != nil {
		slog.Error("failed to get phone number", slog.String("instanceID", instanceID), slog.String("userID", userID), slog.String("error", err.Error()))
		return err
	}

	if phoneInfos.ConfirmedAt < 1 {
		// phone number is not confirmed yet
		slog.Error("phone number is not confirmed", slog.String("instanceID", instanceID), slog.String("userID", userID))
		return errors.New("phone number is not confirmed")
	}

	// generate OTP
	code, err := utils.GenerateOTPCode(OTP_LENGTH)
	if err != nil {
		return err
	}

	// save OTP
	err = pUserDBService.CreateOTP(instanceID, userID, code, userTypes.SMSOTP, MAX_OTP_ATTEMPTS)
	if err != nil {
		return err
	}

	half := len(code) / 2
	formattedCode := fmt.Sprintf("%s-%s", code[:half], code[half:])

	// send SMS
	return sms.SendSMS(
		instanceID, phoneInfos.Phone, userID, sms.SMS_MESSAGE_TYPE_OTP, user.Account.PreferredLanguage, map[string]string{
			"verificationCode": formattedCode,
		},
	)
}

func VerifyOTP(
	instanceID,
	userID,
	code string,
) (*userTypes.OTP, error) {
	otp, err := pUserDBService.FindOTP(instanceID, userID, code)
	if err != nil {
		return nil, err
	}

	if otp.CreatedAt.Before(time.Now().Add(-userDB.OTP_TTL * time.Second)) {
		return nil, errors.New("OTP has expired")
	}

	err = pUserDBService.DeleteOTP(instanceID, userID, code)
	if err != nil {
		return &otp, err
	}

	return &otp, nil
}

func DeleteUser(
	instanceID,
	userID string,
	notifyStudyService func(instanceID string, profiles []string) error,
	sendEmail func(email string) error,
) error {
	// find user
	user, err := pUserDBService.GetUser(instanceID, userID)
	if err != nil {
		return err
	}

	// get all profiles
	profileIDs := make([]string, len(user.Profiles))
	for i, profile := range user.Profiles {
		profileIDs[i] = profile.ID.Hex()
	}

	// notify study service for each profile
	err = notifyStudyService(instanceID, profileIDs)
	if err != nil {
		return err
	}

	// delete all temp tokens
	err = globalInfosDBServices.DeleteAllTempTokenForUser(instanceID, userID, "")
	if err != nil {
		return err
	}

	// delete all renew tokens
	_, err = pUserDBService.DeleteRenewTokensForUser(instanceID, userID)
	if err != nil {
		return err
	}

	// delete account
	err = pUserDBService.DeleteUser(instanceID, userID)
	if err != nil {
		return err
	}

	// notify user
	err = sendEmail(user.Account.AccountID)
	return err
}
