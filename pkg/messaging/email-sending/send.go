package emailsending

import (
	"errors"
	"log/slog"

	messageDB "github.com/case-framework/case-backend/pkg/db/messaging"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

var (
	HttpClient       *httpclient.ClientConfig
	messageDBService *messageDB.MessagingDBService

	GlobalTemplateInfos = map[string]string{}
)

func InitMessageSendingVariables(
	newClientConfig *httpclient.ClientConfig,
	globalTemplateInfos map[string]string,
	mdb *messageDB.MessagingDBService,
) {
	HttpClient = newClientConfig
	GlobalTemplateInfos = globalTemplateInfos
	messageDBService = mdb
}

type SendEmailReq struct {
	To              []string                        `json:"to"`
	Subject         string                          `json:"subject"`
	Content         string                          `json:"content"`
	HighPrio        bool                            `json:"highPrio"`
	HeaderOverrides *messagingTypes.HeaderOverrides `json:"headerOverrides"`
}

func SendOutgoingEmail(
	outgoing *messagingTypes.OutgoingEmail,
) error {
	if HttpClient == nil || HttpClient.RootURL == "" {
		return errors.New("connection to smtp bridge not initialized")
	}

	sendEmailReq := SendEmailReq{
		To:              outgoing.To,
		Subject:         outgoing.Subject,
		Content:         outgoing.Content,
		HighPrio:        outgoing.HighPrio,
		HeaderOverrides: outgoing.HeaderOverrides,
	}
	resp, err := HttpClient.RunHTTPcall("/send-email", sendEmailReq)
	if err == nil && resp != nil {
		errMsg, hasError := resp["error"]
		if hasError {
			err = errors.New(errMsg.(string))
		}
	}
	return err
}

func SendInstantEmailByTemplate(
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

	// send email
	err = SendOutgoingEmail(outgoingEmail)
	if err != nil {
		slog.Debug("error while sending email", slog.String("error", err.Error()))
		_, errS := messageDBService.AddToOutgoingEmails(instanceID, *outgoingEmail)
		if errS != nil {
			slog.Error("failed to save outgoing email", slog.String("error", errS.Error()))
			return errS
		}
		slog.Debug("failed to send email but saved to outgoing", slog.String("error", err.Error()))
		return err
	}

	_, err = messageDBService.AddToSentEmails(instanceID, *outgoingEmail)
	if err != nil {
		slog.Error("failed to save sent email", slog.String("error", err.Error()))
		return err
	}

	return nil
}
