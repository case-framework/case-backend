package pwhash

import (
	"testing"
)

func TestPasswordHashingMethods(t *testing.T) {
	t.Run("try to hash empty string", func(t *testing.T) {
		hPw, err := HashPassword("")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		match, err := ComparePasswordWithHash(hPw, "")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !match {
			t.Error("password should match hashed value")
		}
	})

	t.Run("hash and compare strings", func(t *testing.T) {
		hPw, err := HashPassword("testPassword")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		match, err := ComparePasswordWithHash(hPw, "testPassword")
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
			return
		}
		if !match {
			t.Error("password should match hashed value")
		}
	})
}
