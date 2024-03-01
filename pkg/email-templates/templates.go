package emailtemplates

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"text/template"

	messagingTypes "github.com/case-framework/case-backend/pkg/types/messaging"
)

func GetTemplateTranslation(tDef messagingTypes.EmailTemplate, lang string) messagingTypes.LocalizedTemplate {
	var defaultTranslation messagingTypes.LocalizedTemplate
	for _, tr := range tDef.Translations {
		if tr.Lang == lang {
			return tr
		} else if tr.Lang == tDef.DefaultLanguage {
			defaultTranslation = tr
		}
	}
	return defaultTranslation
}

func ResolveTemplate(tempName string, templateDef string, contentInfos map[string]string) (content string, err error) {
	if strings.TrimSpace(templateDef) == "" {
		return "", errors.New("empty template `" + tempName)
	}
	tmpl, err := template.New(tempName).Parse(templateDef)
	if err != nil {
		err = fmt.Errorf("error when parsing template %s: %v", tempName, err)
		return "", err
	}
	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, contentInfos)
	if err != nil {
		err = fmt.Errorf("error during executing template %s: %v", tempName, err)
		return "", err
	}
	return tpl.String(), nil
}

func CheckAllTranslationsParsable(tempTranslations messagingTypes.EmailTemplate) (err error) {
	if len(tempTranslations.Translations) == 0 {
		return errors.New("error when decoding template `" + tempTranslations.MessageType + "`: translation list is empty")
	}
	for _, templ := range tempTranslations.Translations {
		templateName := tempTranslations.MessageType + templ.Lang
		decodedTemplate, err := base64.StdEncoding.DecodeString(templ.TemplateDef)
		if err != nil {
			err = fmt.Errorf("error when decoding template %s: %v", templateName, err)
			return err
		}
		_, err = ResolveTemplate(
			templateName,
			string(decodedTemplate),
			make(map[string]string),
		)
		if err != nil {
			return errors.New("could not resolve template for `" + templ.Lang + "` - error: " + err.Error())
		}
	}
	return nil
}
