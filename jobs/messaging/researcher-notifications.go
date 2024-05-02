package main

import (
	"log/slog"
	"sync"

	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func handleResearcherNotifications(wg *sync.WaitGroup) {
	defer wg.Done()
	slog.Info("Start handling researcher notifications")

	for _, instanceID := range conf.InstanceIDs {
		counters := InitMessageCounter()
		messageTemplateCache := map[string]messagingTypes.EmailTemplate{}

		studies, err := studyDBService.GetStudies(instanceID, "", false)
		if err != nil {
			slog.Error("Error getting studies", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
			continue
		}

		for _, study := range studies {
			notifications, err := getResearcherMessages(instanceID, study)
			if err != nil {
				slog.Error("Error getting researcher messages", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
				continue
			}

			for _, notification := range notifications {
				// Retrieve the study email template
				templateName := notification.Message.Type + study.Key
				template, ok := messageTemplateCache[templateName]
				if !ok {
					t, err := messagingDBService.GetStudyEmailTemplateByMessageType(instanceID, study.Key, notification.Message.Type)
					if err != nil {
						counters.IncreaseCounter(false)
						slog.Error("Error getting study email template", slog.String("instanceID", instanceID), slog.String("studyKey", study.Key), slog.String("messageType", notification.Message.Type), slog.String("error", err.Error()))
						continue
					}
					messageTemplateCache[templateName] = *t
					template = *t
				}

				payload := map[string]string{
					"studyKey":      study.Key,
					"participantID": notification.Message.ParticipantID,
				}
				for k, v := range notification.Message.Payload {
					payload[k] = v
				}

				subject, content, err := emailsending.GenerateEmailContent(template, "", payload)
				if err != nil {
					counters.IncreaseCounter(false)
					slog.Error("Error generating email content", slog.String("instanceID", instanceID), slog.String("studyKey", study.Key), slog.String("messageType", notification.Message.Type), slog.String("error", err.Error()))
					continue
				}

				outgoingEmail := messagingTypes.OutgoingEmail{
					MessageType:     notification.Message.Type,
					HeaderOverrides: template.HeaderOverrides,
					To:              notification.To,
					Subject:         subject,
					Content:         content,
				}

				_, err = messagingDBService.AddToOutgoingEmails(instanceID, outgoingEmail)
				if err != nil {
					slog.Error("Failed to save outgoing email", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
					counters.IncreaseCounter(false)
					continue
				}
				counters.IncreaseCounter(true)
			}
		}

		counters.Stop()
		slog.Info("Finished handling researcher notifications", slog.String("instanceID", instanceID), slog.Int("success", counters.Success), slog.Int("failed", counters.Failed))
	}

	slog.Info("Finished handling researcher notifications")
}

type ResearcherNotification struct {
	To      []string
	Message studyTypes.StudyMessage
}

func getResearcherMessages(instanceID string, study studyTypes.Study) ([]ResearcherNotification, error) {
	notifications := []ResearcherNotification{}
	messages, err := studyDBService.FindResearcherMessages(instanceID, study.Key)
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		sendTo := []string{}
		for _, subscription := range study.NotificationSubscriptions {
			if subscription.MessageType == "*" || subscription.MessageType == message.Type {
				sendTo = append(sendTo, subscription.Email)
			}
		}

		if len(sendTo) < 1 {
			slog.Debug("No notification subscriptions for current message", slog.String("studyKey", study.Key))
			_, err := studyDBService.DeleteResearcherMessages(instanceID, study.Key, []string{string(message.ID.Hex())})
			if err != nil {
				slog.Error("Error deleting researcher messages", slog.String("instanceID", instanceID), slog.String("studyKey", study.Key), slog.String("error", err.Error()))
			}
			continue
		}

		notifications = append(notifications, ResearcherNotification{
			To:      sendTo,
			Message: message,
		})
	}
	return notifications, nil
}
