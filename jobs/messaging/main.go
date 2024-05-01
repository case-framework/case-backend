package main

import (
	"log/slog"
	"sync"
	"time"

	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
)

const (
	LAST_SEND_ATTEMPT_LOCK_DURATION = 60 * time.Minute

	OUTGOING_EMAILS_BATCH_SIZE = 10

	MAX_FAILED_ATTEMPTS_BEFORE_STOP = 100
)

func main() {
	slog.Info("Starting messaging job")
	start := time.Now()

	var wg sync.WaitGroup

	if conf.RunTasks.ProcessOutgoingEmails {
		wg.Add(1)
		go handleOutgoingMessages(&wg)
	}

	if conf.RunTasks.ScheduleHandler {
		// TODO: handle auto messages
		// TODO: weekly messages handler
		// TODO: newsletter messages handler
	}

	if conf.RunTasks.StudyMessagesHandler {
		// TODO: study messages handler
	}

	if conf.RunTasks.ResearcherMessagesHandler {
		// TODO: researcher messages handler
	}

	wg.Wait()
	slog.Info("Messaging job completed", slog.String("duration", time.Since(start).String()))
}

func handleOutgoingMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	slog.Info("Start handling outgoing messages")

	for _, instanceID := range conf.InstanceIDs {
		slog.Debug("Start handling outgoing messages for instance", slog.String("instanceID", instanceID))
		counters := InitMessageCounter()
		for {
			if counters.Failed > MAX_FAILED_ATTEMPTS_BEFORE_STOP {
				slog.Error("Too many failed attempts, stopping outgoing messages for instance", slog.String("instanceID", instanceID))
				break
			}
			outgoingEmails, err := messagingDBService.GetOutgoingEmailsForSending(
				instanceID,
				time.Now().Add(-LAST_SEND_ATTEMPT_LOCK_DURATION).Unix(),
				false,
				OUTGOING_EMAILS_BATCH_SIZE,
			)
			if err != nil {
				slog.Error("Failed to get outgoing emails for sending", slog.String("error", err.Error()))
				break
			}

			if len(outgoingEmails) == 0 {
				break
			}

			lastFetch := time.Now()

			// Send emails:
			for _, email := range outgoingEmails {
				batchDuration := time.Since(lastFetch)
				if batchDuration >= LAST_SEND_ATTEMPT_LOCK_DURATION {
					slog.Warn("Last batch took too long, breaking", slog.String("duration", batchDuration.String()), slog.String("instanceID", instanceID))
					counters.IncreaseCounter(false)

					err = messagingDBService.ResetLastSendAttemptForOutgoing(instanceID, email.ID.Hex())
					if err != nil {
						slog.Error("Failed to reset last send attempt for outgoing email", slog.String("error", err.Error()))
					}
					continue
				}

				err := emailsending.SendOutgoingEmail(&email)
				if err != nil {
					counters.IncreaseCounter(false)
					slog.Error("Failed to send email", slog.String("instanceID", instanceID), slog.String("messageType", email.MessageType), slog.String("error", err.Error()))

					err = messagingDBService.ResetLastSendAttemptForOutgoing(instanceID, email.ID.Hex())
					if err != nil {
						slog.Error("Failed to reset last send attempt for outgoing email", slog.String("messageType", email.MessageType), slog.String("error", err.Error()))
					}
					continue
				}

				_, err = messagingDBService.AddToSentEmails(instanceID, email)
				if err != nil {
					counters.IncreaseCounter(false)
					slog.Error("Failed to save sent email", slog.String("error", err.Error()))
					continue
				}
				err = messagingDBService.DeleteOutgoingEmail(instanceID, email.ID.Hex())
				if err != nil {
					slog.Error("Failed to delete outgoing email", slog.String("messageType", email.MessageType), slog.String("error", err.Error()))
				}
				counters.IncreaseCounter(true)
			}
		}

		counters.Stop()
		slog.Info("Finished handling outgoing messages for instance", slog.String("instanceID", instanceID), slog.Int64("duration", counters.Duration), slog.Int("success", counters.Success), slog.Int("failed", counters.Failed))
	}
}
