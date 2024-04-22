package emailsending

import (
	"log/slog"

	messageDB "github.com/case-framework/case-backend/pkg/db/messaging"
)

func QueueEmailByTemplate(
	messageDB *messageDB.MessagingDBService,
	instanceID string,
	to []string,
	messageType string,
	studyKey string,
	lang string,
	payload map[string]string,
	useLowPrio bool,
) error {
	outgoingEmail, err := prepOutgoingEmail(
		messageDB,
		instanceID,
		messageType,
		studyKey,
		lang,
		payload,
		to,
		useLowPrio,
	)
	if err != nil {
		return err
	}

	_, err = messageDB.AddToOutgoingEmails(instanceID, *outgoingEmail)
	if err != nil {
		slog.Error("failed to save outgoing email", slog.String("error", err.Error()))
		return err
	}
	return nil
}
