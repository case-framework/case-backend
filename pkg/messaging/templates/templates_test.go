package templates

import (
	"testing"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

func TestTemplateLanguageSelection(t *testing.T) {
	testTemplate := messagingTypes.EmailTemplate{
		MessageType:     "test-type",
		DefaultLanguage: "en",
		Translations: []messagingTypes.LocalizedTemplate{
			{Lang: "en", Subject: "EN"},
			{Lang: "de", Subject: "DE"},
		},
	}

	t.Run("missing target language", func(t *testing.T) {
		translation := GetTemplateTranslation(testTemplate.Translations, "fr", testTemplate.DefaultLanguage)
		if translation.Subject != "EN" {
			t.Errorf("unexpected translation found: %v", translation)
		}
	})

	t.Run("existing target language", func(t *testing.T) {
		translation := GetTemplateTranslation(testTemplate.Translations, "de", testTemplate.DefaultLanguage)
		if translation.Subject != "DE" {
			t.Errorf("unexpected translation found: %v", translation)
		}
	})
}
