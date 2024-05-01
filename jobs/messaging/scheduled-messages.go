package main

import (
	"log/slog"
	"sync"
	"time"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

func handleScheduledMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	slog.Info("Start handling scheduled messages")

	for _, instanceID := range conf.InstanceIDs {
		var mwg sync.WaitGroup

		activeMessages, err := messagingDBService.GetActiveScheduledEmails(instanceID)
		if err != nil {
			slog.Error("Failed to get active scheduled emails", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
			continue
		}
		if len(activeMessages) < 1 {
			continue
		}

		for _, message := range activeMessages {
			mwg.Add(1)
			go generateMessagesForScheduledEmail(&mwg, instanceID, message)

			message.NextTime += message.Period
			var flagNextTimeInPast = false
			for message.NextTime < time.Now().Unix() {
				flagNextTimeInPast = true
				message.NextTime += message.Period
			}
			if flagNextTimeInPast {
				slog.Warn("Next time for sending auto messages was outdated", slog.String("messageID", message.ID.Hex()), slog.String("label", message.Label), slog.Int64("nextTime", message.NextTime))
			}
			if 0 < message.Until && message.Until < message.NextTime {
				slog.Info("Termination date for auto message schedule is reached, schedule will be deleted", slog.String("messageID", message.ID.Hex()), slog.String("label", message.Label))
				err = messagingDBService.DeleteScheduledEmail(instanceID, message.ID.Hex())
				if err != nil {
					slog.Error("Failed to delete scheduled email", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()))
				}
				return
			}
			_, err := messagingDBService.SaveScheduledEmail(instanceID, message)
			if err != nil {
				slog.Error("Failed to save scheduled email", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()))
				continue
			}
		}

		mwg.Wait()
	}
	slog.Info("Finished handling scheduled messages")
}

func generateMessagesForScheduledEmail(wg *sync.WaitGroup, instanceID string, message messagingTypes.ScheduledEmail) {
	defer wg.Done()
	slog.Debug("Start generating messages for scheduled email", slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()), slog.String("label", message.Label))

}
