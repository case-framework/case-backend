package utils

import (
	"net/mail"
	"regexp"
	"strings"
)

func SanitizeEmail(email string) string {
	email = strings.ToLower(email)
	email = strings.Trim(email, " \n\r")
	return email
}

// CheckEmailFormat to check if input string is a correct email address
func CheckEmailFormat(email string) bool {
	if len(email) > 254 {
		return false
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	// additional regex check for correct email format
	emailRule := regexp.MustCompile(`^[a-zA-Z0-9._%+'-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRule.MatchString(email)
}

// BlurEmailAddress transforms an email address to reduce exposed personal info
func BlurEmailAddress(email string) string {
	items := strings.Split(email, "@")
	if len(items) < 1 || len(items[0]) < 1 {
		return "****@**"
	}

	blurredEmail := string([]rune(items[0])[0]) + "****@" + strings.Join(items[1:], "")
	return blurredEmail
}

// CheckPasswordFormat to check if password fulfills password rules
func CheckPasswordFormat(password string) bool {
	pl := len(password)
	if pl < 8 || pl > 512 {
		return false
	}

	var res = 0

	lowercase := regexp.MustCompile("[a-z]")
	uppercase := regexp.MustCompile("[A-Z]")
	number := regexp.MustCompile(`\d`) //"^(?:(?=.*[a-z])(?:(?=.*[A-Z])(?=.*[\\d\\W])|(?=.*\\W)(?=.*\d))|(?=.*\W)(?=.*[A-Z])(?=.*\d)).{8,}$")
	symbol := regexp.MustCompile(`\W`)

	if lowercase.MatchString(password) {
		res++
	}
	if uppercase.MatchString(password) {
		res++
	}
	if number.MatchString(password) {
		res++
	}
	if symbol.MatchString(password) {
		res++
	}
	return res > 2
}

// CheckLanguageCode checks if a string can be considered as a language code
func CheckLanguageCode(code string) bool {
	codeRule := regexp.MustCompile("^[a-z]{2}(-[a-zA-z]{2})?$")
	return codeRule.MatchString(code)
}
