package main

import (
	"log/slog"
	"time"
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

		/*count, err := participantUserDBService.SendReminderToConfirmAccounts(instanceID, createdBefore)
		if err != nil {
			slog.Error("Error sending reminders to confirm accounts", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}*/

		//slog.Info("Sending reminders to confirm accounts finished", slog.String("instanceID", instanceID), slog.Int("count", int(count)))
	}
}
