package emailsending

import (
	"encoding/base64"
	"errors"
	"log/slog"

	messageDB "github.com/case-framework/case-backend/pkg/db/messaging"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	emailtemplates "github.com/case-framework/case-backend/pkg/messaging/email-templates"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

var (
	HttpClient httpclient.ClientConfig

	GlobalTemplateInfos = map[string]string{}
)

func InitMessageSendingVariables(
	newClientConfig httpclient.ClientConfig,
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
	if HttpClient.RootURL == "" {
		return errors.New("email client address not set")
	}

	// get email template
	var templateDef *messagingTypes.EmailTemplate
	var err error
	if studyKey == "" {
		templateDef, err = messageDB.GetGlobalEmailTemplateByMessageType(instanceID, messageType)
	} else {
		templateDef, err = messageDB.GetStudyEmailTemplateByMessageType(instanceID, messageType, studyKey)
	}
	if err != nil {
		return err
	}

	translation := emailtemplates.GetTemplateTranslation(*templateDef, lang)

	decodedTemplate, err := base64.StdEncoding.DecodeString(translation.TemplateDef)
	if err != nil {
		return err
	}

	if payload == nil {
		payload = map[string]string{}
	}
	for k, v := range GlobalTemplateInfos {
		payload[k] = v
	}

	payload["language"] = lang
	// execute template
	templateName := instanceID + messageType + studyKey + lang
	content, err := emailtemplates.ResolveTemplate(
		templateName,
		string(decodedTemplate),
		payload,
	)
	if err != nil {
		return err
	}

	outgoingEmail := messagingTypes.OutgoingEmail{
		MessageType:     messageType,
		To:              to,
		HeaderOverrides: templateDef.HeaderOverrides,
		Subject:         translation.Subject,
		Content:         content,
		HighPrio:        !useLowPrio,
	}

	// send email
	sendEmailReq := SendEmailReq{
		To:              to,
		Subject:         translation.Subject,
		Content:         content,
		HighPrio:        !useLowPrio,
		HeaderOverrides: templateDef.HeaderOverrides,
	}
	_, err = HttpClient.RunHTTPcall("/send-email", sendEmailReq)
	if err != nil {
		slog.Debug("error while sending email", slog.String("error", err.Error()))
		_, errS := messageDB.AddToOutgoingEmails(instanceID, outgoingEmail)
		if errS != nil {
			slog.Error("failed to save outgoing email", slog.String("error", errS.Error()))
			return errS
		}
		slog.Debug("failed to send email but saved to outgoing", slog.String("error", err.Error()))
		return err
	}

	_, err = messageDB.AddToSentEmails(instanceID, outgoingEmail)
	if err != nil {
		slog.Error("failed to save sent email", slog.String("error", err.Error()))
		return err
	}

	return nil
}
