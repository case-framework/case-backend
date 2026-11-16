package utils

import (
	"bufio"
	"log/slog"
	"net/mail"
	"os"
	"regexp"
	"strings"
)

const (
	PASSWORD_MIN_LEN = 12
	PASSWORD_MAX_LEN = 512
)

var blockedPasswords map[string]struct{}

func LoadBlockedPasswords(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	blockedPasswords = make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	lines := 0
	usedEntries := 0
	for scanner.Scan() {
		lines += 1
		passwordEntry := scanner.Text()
		passwordEntry = strings.TrimSpace(passwordEntry)
		passwordEntry = strings.Trim(passwordEntry, "\n")
		if CheckPasswordFormat(passwordEntry) {
			usedEntries += 1
			blockedPasswords[scanner.Text()] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	slog.Info("loaded blocked password list", slog.Int("lines", lines), slog.Int("used", usedEntries))
	return nil
}

func SanitizeEmail(email string) string {
	email = strings.ToLower(email)
	email = strings.Trim(email, " \n\r")
	return email
}

func SanitizePhoneNumber(phone string) string {
	phone = strings.ToLower(phone)
	phone = strings.Trim(phone, " \n\r")
	return phone
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

// BlurEmailAddress generates a short profile alias masking both parts of the
// address. With one rune per part, it uses at most 13 UTF-16 code units.
func BlurEmailAddress(email string) string {
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" || strings.Contains(domain, "@") {
		return "****@****"
	}
	return string([]rune(local)[0]) + "****@" + string([]rune(domain)[0]) + "****"
}

// IsEmailDerivedProfileAlias recognizes both current and legacy defaults when
// an account email changes. Custom aliases should remain unchanged.
func IsEmailDerivedProfileAlias(alias, email string) bool {
	if alias == BlurEmailAddress(email) {
		return true
	}
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" || strings.Contains(domain, "@") {
		return false
	}
	legacy := string([]rune(local)[0]) + "****@" + domain
	return alias == legacy
}

// CheckPasswordFormat to check if password fulfills password rules
func CheckPasswordFormat(password string) bool {
	// Count UTF-16 code units, matching JavaScript string.length.
	pl := 0
	for _, r := range password {
		pl++
		if r > 0xFFFF {
			pl++
		}
		if pl > PASSWORD_MAX_LEN {
			return false
		}
	}
	if pl < PASSWORD_MIN_LEN {
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

func IsPasswordOnBlocklist(password string) bool {
	_, exists := blockedPasswords[password]
	return exists
}

// CheckLanguageCode checks if a string can be considered as a language code
func CheckLanguageCode(code string) bool {
	codeRule := regexp.MustCompile(`^[a-z]{2}(-([A-Z]{2}|\d{3}))?$`)
	return codeRule.MatchString(code)
}
