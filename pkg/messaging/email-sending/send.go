package emailsending

import (
	"errors"
	"log/slog"

	messageDB "github.com/case-framework/case-backend/pkg/db/messaging"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

var (
	HttpClient *httpclient.ClientConfig

	GlobalTemplateInfos = map[string]string{}
)

func InitMessageSendingVariables(
	newClientConfig *httpclient.ClientConfig,
	globalTemplateInfos map[string]string,
) {
	HttpClient = newClientConfig
	GlobalTemplateInfos = globalTemplateInfos
}

type SendEmailReq struct {
	To              []string                        `json:"to"`
	Subject         string                          `json:"subject"`
	Content         string                          `json:"content"`
	HighPrio        bool                            `json:"highPrio"`
	HeaderOverrides *messagingTypes.HeaderOverrides `json:"headerOverrides"`
}

func SendInstantEmailByTemplate(
	messageDB *messageDB.MessagingDBService,
	instanceID string,
	to []string,
	messageType string,
	studyKey string,
	lang string,
	payload map[string]string,
	useLowPrio bool,
) error {
	if HttpClient == nil || HttpClient.RootURL == "" {
		return errors.New("connection to smtp bridge not initialized")
	}

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

	// send email
	sendEmailReq := SendEmailReq{
		To:              to,
		Subject:         outgoingEmail.Subject,
		Content:         outgoingEmail.Content,
		HighPrio:        outgoingEmail.HighPrio,
		HeaderOverrides: outgoingEmail.HeaderOverrides,
	}
	_, err = HttpClient.RunHTTPcall("/send-email", sendEmailReq)
	if err != nil {
		slog.Debug("error while sending email", slog.String("error", err.Error()))
		_, errS := messageDB.AddToOutgoingEmails(instanceID, *outgoingEmail)
		if errS != nil {
			slog.Error("failed to save outgoing email", slog.String("error", errS.Error()))
			return errS
		}
		slog.Debug("failed to send email but saved to outgoing", slog.String("error", err.Error()))
		return err
	}

	_, err = messageDB.AddToSentEmails(instanceID, *outgoingEmail)
	if err != nil {
		slog.Error("failed to save sent email", slog.String("error", err.Error()))
		return err
	}

	return nil
}
