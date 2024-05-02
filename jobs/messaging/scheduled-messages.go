package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	umTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
	"go.mongodb.org/mongo-driver/bson"
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

	switch message.Type {
	case "all-users":
		generateScheduledEmailsForAllUsers(
			instanceID,
			message,
		)
	case "study-participants":
		message.Template.StudyKey = message.StudyKey
		generateScheduledEmailsForStudyParticipants(
			instanceID,
			message,
		)
	default:
		slog.Error("message schedule type unknown", slog.String("type", message.Type), slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()))
	}
}

func generateScheduledEmailsForAllUsers(instanceID string, message messagingTypes.ScheduledEmail) {
	counters := InitMessageCounter()

	filter := bson.M{
		"account.accountConfirmedAt":                       bson.M{"$gt": 0},
		"contactPreferences.receiveWeeklyMessageDayOfWeek": time.Now().Weekday(),
	}

	err := participantUserDBService.FindAndExecuteOnUsers(
		context.Background(),
		instanceID,
		filter,
		nil,
		false,
		func(user umTypes.User, args ...interface{}) error {
			if !isSubscribed(&user, message.Template.MessageType) {
				return nil
			}

			if !hasAccountType(&user, "email") {
				return nil
			}

			outgoingEmail, err := prepOutgoingFromScheduledEmail(
				instanceID,
				message,
				user,
			)
			if err != nil {
				slog.Error("Failed to prepare outgoing email", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()), slog.String("userID", user.ID.Hex()))
				counters.IncreaseCounter(false)
				return err
			}

			_, err = messagingDBService.AddToOutgoingEmails(instanceID, *outgoingEmail)
			if err != nil {
				slog.Error("Failed to save outgoing email", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()), slog.String("userID", user.ID.Hex()))
				counters.IncreaseCounter(false)
				return err
			}

			counters.IncreaseCounter(true)
			return nil
		},
	)
	counters.Stop()
	if err != nil {
		slog.Error("Failed to get users for sending scheduled email", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()), slog.Int("generatedMessages", counters.Success), slog.Int("failedMessages", counters.Failed))
		return
	}
	slog.Info("Generated messages for scheduled email", slog.String("instanceID", instanceID), slog.String("messageID", message.ID.Hex()), slog.Int("generatedMessages", counters.Success), slog.Int("failedMessages", counters.Failed))
}

func generateScheduledEmailsForStudyParticipants(instanceID string, message messagingTypes.ScheduledEmail) {
	// TODO:
}

func isSubscribed(user *umTypes.User, messageType string) bool {
	switch messageType {
	case messagingTypes.EMAIL_TYPE_WEEKLY:
		return user.ContactPreferences.SubscribedToWeekly
	case messagingTypes.EMAIL_TYPE_NEWSLETTER:
		return user.ContactPreferences.SubscribedToNewsletter
	}
	return true
}

func hasAccountType(user *umTypes.User, accountType string) bool {
	return user.Account.Type == accountType
}

func prepOutgoingFromScheduledEmail(
	instanceID string,
	message messagingTypes.ScheduledEmail,
	user umTypes.User,
) (*messagingTypes.OutgoingEmail, error) {
	outgoingEmail := messagingTypes.OutgoingEmail{
		MessageType:     message.Template.MessageType,
		HeaderOverrides: message.Template.HeaderOverrides,
	}

	if user.Account.Type == "email" {
		outgoingEmail.To = []string{user.Account.AccountID}
	}

	payload := map[string]string{}
	for k, v := range emailsending.GlobalTemplateInfos {
		payload[k] = v
	}
	payload["language"] = user.Account.PreferredLanguage

	if message.Template.MessageType == messagingTypes.EMAIL_TYPE_NEWSLETTER {
		outgoingEmail.To = getEmailsByIds(user.ContactInfos, user.ContactPreferences.SendNewsletterTo)
		token, err := getUnsubscribeToken(instanceID, user)
		if err != nil {
			return nil, err
		}
		payload["unsubscribeToken"] = token
	} else {
		token, err := getTemploginToken(instanceID, user, message.Template.StudyKey)
		if err != nil {
			return nil, err
		}
		payload["loginToken"] = token
	}

	payload["studyKey"] = message.StudyKey

	subject, content, err := emailsending.GenerateEmailContent(message.Template, user.Account.PreferredLanguage, payload)
	if err != nil {
		return nil, err
	}

	outgoingEmail.Subject = subject
	outgoingEmail.Content = content
	return &outgoingEmail, nil
}

func getEmailsByIds(contacts []umTypes.ContactInfo, ids []string) []string {
	emails := []string{}
	for _, c := range contacts {
		if c.Type == "email" {
			for _, id := range ids {
				if c.ID.Hex() == id && c.ConfirmedAt > 0 {
					emails = append(emails, c.Email)
				}
			}
		}
	}
	return emails
}

func getTemploginToken(instanceID string, user umTypes.User, studyKey string) (string, error) {
	tempTokenInfos := umTypes.TempToken{
		UserID:     user.ID.Hex(),
		InstanceID: instanceID,
		Purpose:    umTypes.TOKEN_PURPOSE_SURVEY_LOGIN,
		Info:       map[string]string{"studyKey": studyKey},
		Expiration: umUtils.GetExpirationTime(conf.Intervals.LoginTokenTTL),
	}
	tempToken, err := globalInfosDBService.AddTempToken(tempTokenInfos)
	if err != nil {
		slog.Error("failed to create login token", slog.String("error", err.Error()))
		return "", err
	}

	return tempToken, nil
}

func getUnsubscribeToken(instanceID string, user umTypes.User) (string, error) {
	tempTokenInfos := umTypes.TempToken{
		UserID:     user.ID.Hex(),
		InstanceID: instanceID,
		Purpose:    umTypes.TOKEN_PURPOSE_UNSUBSCRIBE_NEWSLETTER,
		Info:       nil,
		Expiration: umUtils.GetExpirationTime(conf.Intervals.UnsubscribeTokenTTL),
	}
	tempToken, err := globalInfosDBService.AddTempToken(tempTokenInfos)
	if err != nil {
		slog.Error("failed to create unsubscribe token", slog.String("error", err.Error()))
		return "", err
	}

	return tempToken, nil
}
