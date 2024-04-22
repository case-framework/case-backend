package main

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	emailTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	umTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
)

func main() {
	slog.Info("Starting user management job")
	start := time.Now()

	cleanUpUnverifiedUsers()
	sendReminderToConfirmAccounts()

	// TODO: detect and notify inactive users
	// TODO: clean up users marked for deletion

	slog.Info("User management jobs completed", slog.Duration("duration", time.Since(start)))
}

func cleanUpUnverifiedUsers() {
	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start cleaning up unverified users", slog.String("instanceID", instanceID))

		// call DB method participantUserDBService
		createdBefore := time.Now().Add(-conf.UserManagementConfig.DeleteUnverifiedUsersAfter).Unix()
		count, err := participantUserDBService.DeleteUnverifiedUsers(instanceID, createdBefore)
		if err != nil {
			slog.Error("Error cleaning up unverified users", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		slog.Info("Clean up unverified users finished", slog.String("instanceID", instanceID), slog.Int("count", int(count)))

	}

}

func sendReminderToConfirmAccounts() {
	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start sending reminders to confirm accounts", slog.String("instanceID", instanceID))

		createdBefore := time.Now().Add(-conf.UserManagementConfig.SendReminderToConfirmAccountAfter).Unix()
		filter := bson.M{}
		filter["$and"] = bson.A{
			bson.M{"account.accountConfirmedAt": bson.M{"$lt": 1}},
			bson.M{"timestamps.reminderToConfirmSentAt": bson.M{"$lt": 1}},
			bson.M{"timestamps.createdAt": bson.M{"$lt": createdBefore}},
		}

		count := 0

		// call DB method participantUserDBService
		err := participantUserDBService.FindAndExecuteOnUsers(
			context.Background(),
			instanceID,
			filter,
			nil,
			false,
			func(user umTypes.User, args ...interface{}) error {
				// Generate token
				tempTokenInfos := umTypes.TempToken{
					UserID:     user.ID.Hex(),
					InstanceID: instanceID,
					Purpose:    umTypes.TOKEN_PURPOSE_CONTACT_VERIFICATION,
					Info: map[string]string{
						"type":  umTypes.ACCOUNT_TYPE_EMAIL,
						"email": user.Account.AccountID,
					},
					Expiration: umUtils.GetExpirationTime(conf.UserManagementConfig.EmailContactVerificationTokenTTL),
				}
				tempToken, err := globalInfosDBService.AddTempToken(tempTokenInfos)
				if err != nil {
					slog.Error("failed to create verification token", slog.String("error", err.Error()))
					return err
				}

				// Call message sending
				err = emailsending.SendInstantEmailByTemplate(
					messagingDBService,
					instanceID,
					[]string{
						user.Account.AccountID,
					},
					emailTypes.EMAIL_TYPE_REGISTRATION,
					"",
					user.Account.PreferredLanguage,
					map[string]string{
						"token": tempToken,
					},
					true,
				)
				if err != nil {
					slog.Error("failed to send verification email", slog.String("error", err.Error()))
					return err
				}

				// Update user record
				update := bson.M{"$set": bson.M{"timestamps.reminderToConfirmSentAt": time.Now().Unix()}}
				err = participantUserDBService.UpdateUser(instanceID, user.ID.Hex(), update)
				if err != nil {
					slog.Error("failed to update user record", slog.String("error", err.Error()))
					return err
				}

				count = count + 1
				return nil
			},
		)
		if err != nil {
			slog.Error("Error sending reminders to confirm accounts", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		slog.Info("Sending reminders to confirm accounts finished", slog.String("instanceID", instanceID), slog.Int("count", int(count)))
	}
}
