package utils

import "regexp"

// check if a string value can be safely used as a part of an URL
func IsURLSafe(value string) bool {
	if value == "" {
		return false
	}

	pattern := `^[a-zA-Z0-9-_]+$`
	regex := regexp.MustCompile(pattern)

	return regex.MatchString(value)
}
