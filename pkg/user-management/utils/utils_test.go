package utils

import (
	"testing"
)

func TestSanitizeEmail(t *testing.T) {
	t.Run("with different formats", func(t *testing.T) {
		email := SanitizeEmail("\n23234@test.DE")
		if email != "23234@test.de" {
			t.Errorf("unexpected email: %s", email)
		}

		email = SanitizeEmail("  \n 23234@test.DE \n\r")
		if email != "23234@test.de" {
			t.Errorf("unexpected email: %s", email)
		}

		email = SanitizeEmail("23234@test.de")
		if email != "23234@test.de" {
			t.Errorf("unexpected email: %s", email)
		}
	})
}

func TestBlurEmailAddress(t *testing.T) {
	t.Run("with different formats", func(t *testing.T) {
		email := BlurEmailAddress("a@test.de")
		if email != "a****@test.de" {
			t.Errorf("unexpected email: %s", email)
		}

		email = BlurEmailAddress("a1234@test.de")
		if email != "a****@test.de" {
			t.Errorf("unexpected email: %s", email)
		}

		email = BlurEmailAddress("a123sdfsdfsdfa34@test.de")
		if email != "a****@test.de" {
			t.Errorf("unexpected email: %s", email)
		}
	})
}

func TestCheckPasswordFormat(t *testing.T) {
	t.Run("with a too short password", func(t *testing.T) {
		if CheckPasswordFormat("1n34T6@") {
			t.Error("should be false")
		}
	})
	t.Run("with a too weak password", func(t *testing.T) {
		if CheckPasswordFormat("13342678") {
			t.Error("should be false")
		}
		if CheckPasswordFormat("11111aaaa") {
			t.Error("should be false")
		}
	})
	t.Run("with good passwords", func(t *testing.T) {

		if !CheckPasswordFormat("1n34T678") {
			t.Error("should be true")
		}
		if !CheckPasswordFormat("nnnnnnT@@") {
			t.Error("should be true")
		}
		if !CheckPasswordFormat("TTTTTTTT77.") {
			t.Error("should be true")
		}
		if !CheckPasswordFormat("Tt1,.Lo%4") {
			t.Error("should be true")
		}
	})
}

func TestCheckEmailFormat(t *testing.T) {
	t.Run("with missing @", func(t *testing.T) {
		if CheckEmailFormat("t.t.com") {
			t.Error("should be false")
		}
	})

	t.Run("with wrong domain format", func(t *testing.T) {
		if CheckEmailFormat("t@t.") {
			t.Error("should be false")
		}
	})

	t.Run("with missing top level domain", func(t *testing.T) {
		if CheckEmailFormat("t@com") {
			t.Error("should be false")
		}
	})

	t.Run("with wrong local format", func(t *testing.T) {
		if CheckEmailFormat("@t.com") {
			t.Error("should be false")
		}
	})

	t.Run("with too many @", func(t *testing.T) {
		if CheckEmailFormat("t@@t.com") {
			t.Error("should be false")
		}
	})

	t.Run("with ..", func(t *testing.T) {
		if CheckEmailFormat("t..t@t.com") {
			t.Error("should be false")
		}
	})

	t.Run("with correct format", func(t *testing.T) {
		if !CheckEmailFormat("t@t.com") {
			t.Error("should be true")
		}
	})

	t.Run("with correct format", func(t *testing.T) {
		if !CheckEmailFormat("t+1@t.com") {
			t.Error("should be true")
		}
	})
}

func TestLanguageCodeFormat(t *testing.T) {
	t.Run("with t", func(t *testing.T) {
		if CheckLanguageCode("t") {
			t.Error("should be false")
		}
	})

	t.Run("with ttt", func(t *testing.T) {
		if CheckLanguageCode("ttt") {
			t.Error("should be false")
		}
	})

	t.Run("with 1t", func(t *testing.T) {
		if CheckLanguageCode("1t") {
			t.Error("should be false")
		}
	})

	t.Run("with TT", func(t *testing.T) {
		if CheckLanguageCode("TT") {
			t.Error("should be false")
		}
	})

	t.Run("with .t", func(t *testing.T) {
		if CheckLanguageCode(".t") {
			t.Error("should be false")
		}
	})

	t.Run("with tt", func(t *testing.T) {
		if !CheckLanguageCode("tt") {
			t.Error("should be true")
		}
	})

}
