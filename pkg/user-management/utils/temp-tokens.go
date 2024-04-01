package utils

import (
	"crypto/rand"
	"strings"
	"time"

	b32 "encoding/base32"
)

func GenerateUniqueTokenString() (string, error) {
	t := time.Now()
	ms := uint64(t.Unix())*1000 + uint64(t.Nanosecond()/int(time.Millisecond))

	token := make([]byte, 24)
	token[0] = byte(ms >> 40)
	token[1] = byte(ms >> 32)
	token[2] = byte(ms >> 24)
	token[3] = byte(ms >> 16)
	token[4] = byte(ms >> 8)
	token[5] = byte(ms)

	_, err := rand.Read(token[6:])
	if err != nil {
		return "", err
	}

	tokenStr := b32.StdEncoding.WithPadding(b32.NoPadding).EncodeToString(token)
	tokenStr = strings.ToLower(tokenStr)
	return tokenStr, nil
}

func GetExpirationTime(validityPeriod time.Duration) time.Time {
	return time.Now().Add(validityPeriod)
}

func ReachedExpirationTime(t time.Time) bool {
	return time.Now().After(t)
}
