package usermanagement

import (
	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
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
