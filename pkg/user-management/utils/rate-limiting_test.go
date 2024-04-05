package utils

import (
	"testing"
	"time"
)

func TestHasMoreAttemptsRecently(t *testing.T) {
	t.Run("with a empty array", func(t *testing.T) {
		if HasMoreAttemptsRecently([]int64{}, 5, 100) {
			t.Error("should be false")
		}
	})

	t.Run("with less than threshold", func(t *testing.T) {
		if HasMoreAttemptsRecently([]int64{
			time.Now().Unix() - 105, time.Now().Unix() - 65, time.Now().Unix() - 55, time.Now().Unix() - 45, time.Now().Unix() - 35, time.Now().Unix() - 25,
		}, 5, 100) {
			t.Error("should be false")
		}
	})

	t.Run("with more than threshold", func(t *testing.T) {
		if !HasMoreAttemptsRecently([]int64{
			time.Now().Unix() - 95, time.Now().Unix() - 65, time.Now().Unix() - 55, time.Now().Unix() - 45, time.Now().Unix() - 35, time.Now().Unix() - 25,
		}, 5, 100) {
			t.Error("should be true")
		}
	})

}
