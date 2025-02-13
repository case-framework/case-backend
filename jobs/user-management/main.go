package main

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	emailTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	studyService "github.com/case-framework/case-backend/pkg/study"
	usermanagement "github.com/case-framework/case-backend/pkg/user-management"
	umTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
)

func main() {
	slog.Info("Starting user management job")
	start := time.Now()

	if conf.RunTasks.CleanUpUnverifiedUsers {
		cleanUpUnverifiedUsers()
	}
	if conf.RunTasks.SendReminderToConfirmAccounts {
		sendReminderToConfirmAccounts()
	}
	if conf.RunTasks.HandleInactiveUsers {
		notifyInactiveUsersAndMarkForDeletion()
		cleanUpUsersMarkedForDeletion()
	}

	if conf.RunTasks.GenerateProfileIDLookup {
		generateProfileIDLookup()
	}

	slog.Info("User management jobs completed", slog.String("duration", time.Since(start).String()))
}

func cleanUpUnverifiedUsers() {
	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start cleaning up unverified users", slog.String("instanceID", instanceID))

		count := 0
		// call DB method participantUserDBService
		createdBefore := time.Now().Add(-conf.UserManagementConfig.DeleteUnverifiedUsersAfter).Unix()
		filter := bson.M{}
		filter["$and"] = bson.A{
			bson.M{"account.accountConfirmedAt": 0},
			bson.M{"timestamps.createdAt": bson.M{"$lt": createdBefore}},
		}
		err := participantUserDBService.FindAndExecuteOnUsers(
			context.Background(),
			instanceID,
			filter,
			nil,
			false,
			func(user umTypes.User, args ...interface{}) error {
				err := usermanagement.DeleteUser(
					instanceID,
					user.ID.Hex(),
					func(instanceID string, profiles []string) error {
						for _, profile := range profiles {
							studyService.OnProfileDeleted(instanceID, profile, nil)
						}
						return nil
					},
					func(email string) error {
						err := emailsending.QueueEmailByTemplate(
							instanceID,
							[]string{
								email,
							},
							emailTypes.EMAIL_TYPE_ACCOUNT_DELETED,
							"",
							user.Account.PreferredLanguage,
							map[string]string{},
							true,
						)
						if err != nil {
							slog.Error("failed to queue account deleted email", slog.String("error", err.Error()))
							return err
						}
						return nil
					},
				)
				if err != nil {
					slog.Error("failed to delete user", slog.String("error", err.Error()))
					return err
				}
				count = count + 1
				return nil
			},
		)
		if err != nil {
			slog.Error("Error cleaning up unverified users", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		slog.Info("Clean up unverified users finished", slog.String("instanceID", instanceID), slog.Int("count", int(count)))
	}
}

func sendReminderToConfirmAccounts() {
	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start preparing reminders to confirm accounts", slog.String("instanceID", instanceID))

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
				err = emailsending.QueueEmailByTemplate(
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
					slog.Error("failed to queue verification email", slog.String("error", err.Error()))
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

		slog.Info("Preparing reminders to confirm accounts finished", slog.String("instanceID", instanceID), slog.Int("count", int(count)))
	}
}

func notifyInactiveUsersAndMarkForDeletion() {
	if conf.UserManagementConfig.NotifyAfterInactiveFor == 0 {
		slog.Info("Inactive user notification is disabled")
		return
	}

	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start notifying inactive users and mark for deletion", slog.String("instanceID", instanceID))

		count := 0

		lastActivityEarlierThan := time.Now().Add(-conf.UserManagementConfig.NotifyAfterInactiveFor).Unix()
		filter := bson.M{}
		filter["$and"] = bson.A{
			bson.M{
				"roles": bson.M{"$nin": bson.A{
					"SERVICE",
					"RESEARCHER",
					"ADMIN",
				}},
			}, // for legacy reasons
			bson.M{"timestamps.lastLogin": bson.M{"$lt": lastActivityEarlierThan}},
			bson.M{"timestamps.lastTokenRefresh": bson.M{"$lt": lastActivityEarlierThan}},
			bson.M{"timestamps.markedForDeletion": bson.M{"$not": bson.M{"$gt": 0}}},
		}

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
					Purpose:    umTypes.TOKEN_PURPOSE_INACTIVE_USER_NOTIFICATION,
					Info: map[string]string{
						"type":  umTypes.ACCOUNT_TYPE_EMAIL,
						"email": user.Account.AccountID,
					},
					Expiration: umUtils.GetExpirationTime(conf.UserManagementConfig.MarkForDeletionAfterInactivityNotification),
				}
				tempToken, err := globalInfosDBService.AddTempToken(tempTokenInfos)
				if err != nil {
					slog.Error("failed to create verification token", slog.String("error", err.Error()))
					return err
				}

				// Call message sending
				err = emailsending.QueueEmailByTemplate(
					instanceID,
					[]string{
						user.Account.AccountID,
					},
					emailTypes.EMAIL_TYPE_ACCOUNT_INACTIVITY,
					"",
					user.Account.PreferredLanguage,
					map[string]string{
						"token": tempToken,
					},
					true,
				)
				if err != nil {
					slog.Error("failed to queue inactivity notice email", slog.String("error", err.Error()))
					return err
				}

				// Update user record
				update := bson.M{"$set": bson.M{"timestamps.markedForDeletion": time.Now().Add(conf.UserManagementConfig.MarkForDeletionAfterInactivityNotification).Unix()}}
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
			slog.Error("Error notifying inactive users and mark for deletion", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		slog.Info("Notifying inactive users and mark for deletion finished", slog.String("instanceID", instanceID), slog.Int("count", int(count)))
	}
}

func cleanUpUsersMarkedForDeletion() {
	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start cleaning up users marked for deletion", slog.String("instanceID", instanceID))

		count := 0

		// call DB method participantUserDBService
		filter := bson.M{}
		filter["$and"] = bson.A{
			bson.M{"timestamps.markedForDeletion": bson.M{"$gt": 0}},
			bson.M{"timestamps.markedForDeletion": bson.M{"$lt": time.Now().Unix()}},
		}
		err := participantUserDBService.FindAndExecuteOnUsers(
			context.Background(),
			instanceID,
			filter,
			nil,
			false,
			func(user umTypes.User, args ...interface{}) error {
				err := usermanagement.DeleteUser(
					instanceID,
					user.ID.Hex(),
					func(instanceID string, profiles []string) error {
						for _, profile := range profiles {
							studyService.OnProfileDeleted(instanceID, profile, nil)
						}
						return nil
					},
					func(email string) error {
						err := emailsending.QueueEmailByTemplate(
							instanceID,
							[]string{
								email,
							},
							emailTypes.EMAIL_TYPE_ACCOUNT_DELETED_AFTER_INACTIVITY,
							"",
							user.Account.PreferredLanguage,
							map[string]string{},
							true,
						)
						if err != nil {
							slog.Error("failed to queue account deleted email", slog.String("error", err.Error()))
							return err
						}
						return nil
					},
				)
				if err != nil {
					slog.Error("failed to delete user", slog.String("error", err.Error()))
					return err
				}
				count = count + 1
				return nil
			},
		)
		if err != nil {
			slog.Error("Error cleaning up users marked for deletion", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		slog.Info("Clean up users marked for deletion finished", slog.String("instanceID", instanceID), slog.Int("count", int(count)))
	}
}
