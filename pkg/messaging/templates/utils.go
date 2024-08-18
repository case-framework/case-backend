package templates

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"strings"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

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

func GetTemplateTranslation(translations []messagingTypes.LocalizedTemplate, lang string, defaultLang string) messagingTypes.LocalizedTemplate {
	var defaultTranslation messagingTypes.LocalizedTemplate
	for _, tr := range translations {
		if tr.Lang == lang {
			return tr
		} else if tr.Lang == defaultLang {
			defaultTranslation = tr
		}
	}
	return defaultTranslation
}

func CheckAllTranslationsParsable(tempTranslations []messagingTypes.LocalizedTemplate, messageType string) error {
	if len(tempTranslations) == 0 {
		return errors.New("error when decoding template: translation list is empty")
	}
	for _, templ := range tempTranslations {
		templateName := messageType + templ.Lang
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
