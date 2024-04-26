package emailsending

import (
	"log/slog"
)

func QueueEmailByTemplate(
	instanceID string,
	to []string,
	messageType string,
	studyKey string,
	lang string,
	payload map[string]string,
	useLowPrio bool,
) error {
	outgoingEmail, err := prepOutgoingEmail(
		messageDBService,
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

	_, err = messageDBService.AddToOutgoingEmails(instanceID, *outgoingEmail)
	if err != nil {
		slog.Error("failed to save outgoing email", slog.String("error", err.Error()))
		return err
	}
	return nil
}
