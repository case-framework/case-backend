package sms

import (
	"encoding/base64"
	"time"

	messageDB "github.com/case-framework/case-backend/pkg/db/messaging"
	"github.com/case-framework/case-backend/pkg/messaging/templates"
	"github.com/case-framework/case-backend/pkg/messaging/types"
)

var (
	SmsGatewayConfig *types.SMSGatewayConfig
	MessageDBService *messageDB.MessagingDBService
)

const (
	SMS_MESSAGE_TYPE_VERIFY_PHONE_NUMBER = "verify-phone-number"
	SMS_MESSAGE_TYPE_OTP                 = "otp"
)

func Init(
	smsGatewayConfig *types.SMSGatewayConfig,
	mdb *messageDB.MessagingDBService,
) {
	SmsGatewayConfig = smsGatewayConfig
	MessageDBService = mdb
}

func SendSMS(instanceID string, to string, userID string, messageType string, lang string, payload map[string]string) error {
	templateDef, err := MessageDBService.GetSMSTemplateByType(instanceID, messageType)
	if err != nil {
		return err
	}

	translation := templates.GetTemplateTranslation(templateDef.Translations, lang, templateDef.DefaultLanguage)

	decodedTemplate, err := base64.StdEncoding.DecodeString(translation.TemplateDef)
	if err != nil {
		return err
	}

	if payload == nil {
		payload = map[string]string{}
	}

	payload["language"] = lang

	// execute template
	templateName := instanceID + messageType + lang
	content, err := templates.ResolveTemplate(
		templateName,
		string(decodedTemplate),
		payload,
	)
	if err != nil {
		return err
	}

	// send sms
	err = runSMSsending(to, content, templateDef.From)
	if err != nil {
		return err
	}

	// save sent sms
	_, err = MessageDBService.AddToSentSMS(instanceID, types.SentSMS{
		MessageType: messageType,
		PhoneNumber: to,
		UserID:      userID,
		SentAt:      time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}
