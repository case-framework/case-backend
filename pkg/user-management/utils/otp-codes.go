package utils

import "crypto/rand"

const codeCharSet = "1234567890"

// GenerateOTPCode generates a random OTP code of the given length
func GenerateOTPCode(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	charsetLength := len(codeCharSet)
	for i := 0; i < length; i++ {
		buffer[i] = codeCharSet[int(buffer[i])%charsetLength]
	}
	return string(buffer), nil
}
