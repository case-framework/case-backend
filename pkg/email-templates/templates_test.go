package emailtemplates

import (
	"testing"

	messagingTypes "github.com/case-framework/case-backend/pkg/types/messaging"
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
		translation := GetTemplateTranslation(testTemplate, "fr")
		if translation.Subject != "EN" {
			t.Errorf("unexpected translation found: %v", translation)
		}
	})

	t.Run("existing target language", func(t *testing.T) {
		translation := GetTemplateTranslation(testTemplate, "de")
		if translation.Subject != "DE" {
			t.Errorf("unexpected translation found: %v", translation)
		}
	})
}

func TestResolveTemplate(t *testing.T) {
	contentInfos := map[string]string{
		"testKey1": "value1",
		"testKey2": "value2",
	}
	t.Run("with static template", func(t *testing.T) {
		content, err := ResolveTemplate("testTemp1", "<h1>Test</h1>", contentInfos)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if content != "<h1>Test</h1>" {
			t.Errorf("unexpected content: %s", content)
		}
	})

	t.Run("with dynamic template missing info", func(t *testing.T) {
		content, err := ResolveTemplate("testTemp2", "<h1>{{index . \"testKey3\"}}</h1>", contentInfos)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if content != "<h1></h1>" {
			t.Errorf("unexpected content: %s", content)
		}
	})

	t.Run("with dynamic template valid infos", func(t *testing.T) {
		content, err := ResolveTemplate("testTemp3", `<h1>{{index . "testKey1"}}</h1>`, contentInfos)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if content != "<h1>value1</h1>" {
			t.Errorf("unexpected content: %s", content)
		}
	})
}
