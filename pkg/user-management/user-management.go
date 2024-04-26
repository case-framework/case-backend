package usermanagement

import (
	"errors"
	"fmt"
	"log/slog"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
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
	sendEmail func(email string, code string, preferredLang string) error,
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
	err = pUserDBService.CreateOTP(instanceID, userID, code, userTypes.EmailOTP)
	if err != nil {
		return err
	}

	half := len(code) / 2
	formattedCode := fmt.Sprintf("%s-%s", code[:half], code[half:])

	// send OTP
	err = sendEmail(user.Account.AccountID, formattedCode, user.Account.PreferredLanguage)
	if err != nil {
		return err
	}

	return nil
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
