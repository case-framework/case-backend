package emailsending

import (
	"encoding/base64"

	messageDB "github.com/case-framework/case-backend/pkg/db/messaging"
	emailtemplates "github.com/case-framework/case-backend/pkg/messaging/email-templates"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

func prepOutgoingEmail(
	messageDB *messageDB.MessagingDBService,
	instanceID string,
	messageType string,
	studyKey string,
	lang string,
	payload map[string]string,
	to []string,
	useLowPrio bool,

) (*messagingTypes.OutgoingEmail, error) {

	// get email template
	var templateDef *messagingTypes.EmailTemplate
	var err error
	if studyKey == "" {
		templateDef, err = messageDB.GetGlobalEmailTemplateByMessageType(instanceID, messageType)
	} else {
		templateDef, err = messageDB.GetStudyEmailTemplateByMessageType(instanceID, messageType, studyKey)
	}
	if err != nil {
		return nil, err
	}

	translation := emailtemplates.GetTemplateTranslation(*templateDef, lang)

	decodedTemplate, err := base64.StdEncoding.DecodeString(translation.TemplateDef)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	outgoingEmail := messagingTypes.OutgoingEmail{
		MessageType:     messageType,
		To:              to,
		HeaderOverrides: templateDef.HeaderOverrides,
		Subject:         translation.Subject,
		Content:         content,
		HighPrio:        !useLowPrio,
	}
	return &outgoingEmail, nil
}

func GenerateEmailContent(
	templateDef messagingTypes.EmailTemplate,
	lang string,
	payload map[string]string,
) (string, string, error) {
	translation := emailtemplates.GetTemplateTranslation(templateDef, lang)

	decodedTemplate, err := base64.StdEncoding.DecodeString(translation.TemplateDef)
	if err != nil {
		return "", "", err
	}

	if payload == nil {
		payload = map[string]string{}
	}
	for k, v := range GlobalTemplateInfos {
		payload[k] = v
	}

	// execute template
	templateName := templateDef.ID.Hex() + lang
	content, err := emailtemplates.ResolveTemplate(
		templateName,
		string(decodedTemplate),
		payload,
	)
	if err != nil {
		return "", "", err
	}

	return translation.Subject, content, nil
}
