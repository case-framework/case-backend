package emailtemplates

import (
	"github.com/case-framework/case-backend/pkg/messaging/templates"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

func GetTemplateTranslation(tDef messagingTypes.EmailTemplate, lang string) messagingTypes.LocalizedTemplate {
	return templates.GetTemplateTranslation(tDef.Translations, lang, tDef.DefaultLanguage)
}

func CheckAllTranslationsParsable(tempTranslations messagingTypes.EmailTemplate) (err error) {
	return templates.CheckAllTranslationsParsable(tempTranslations.Translations, tempTranslations.MessageType)
}
