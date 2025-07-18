package utils

import (
	"os"
	"regexp"
	"strings"
)

// GenerateEnvVarName generates a standardized environment variable name from a given string.
// It converts the input to uppercase and replaces any non-alphanumeric characters with underscores.
// Leading and trailing underscores are removed.
func GenerateEnvVarName(input string) string {
	// Convert to uppercase
	normalized := strings.ToUpper(input)

	// Replace any non-alphanumeric characters with underscores
	reg := regexp.MustCompile(`[^A-Z0-9]+`)
	normalized = reg.ReplaceAllString(normalized, "_")

	// Remove leading/trailing underscores
	normalized = strings.Trim(normalized, "_")

	return normalized
}

// GenerateExternalServiceAPIKeyEnvVarName generates an environment variable name for an external service's API key
// based on its name. Format: EXTERNAL_SERVICE_API_KEY_FOR_{NORMALIZED_NAME}
func GenerateExternalServiceAPIKeyEnvVarName(serviceName string) string {
	normalizedName := GenerateEnvVarName(serviceName)
	return "EXTERNAL_SERVICE_API_KEY_FOR_" + normalizedName
}

// GenerateConfidentialResponseExportSecretEnvVarName generates an environment variable name for a confidential
// export task's study global secret based on its name. Format: CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_{NORMALIZED_NAME}
func GenerateConfidentialResponseExportSecretEnvVarName(taskName string) string {
	normalizedName := GenerateEnvVarName(taskName)
	return "CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_" + normalizedName
}

// GenerateSmtpServerUsernameEnvVarName generates an environment variable name for an SMTP server's username
// based on its hostname and port. Format: SMTP_SERVER_USERNAME_FOR_{NORMALIZED_HOST}_{NORMALIZED_PORT}
func GenerateSmtpServerUsernameEnvVarName(hostname, port string) string {
	normalizedHost := GenerateEnvVarName(hostname)
	normalizedPort := GenerateEnvVarName(port)
	return "SMTP_SERVER_USERNAME_FOR_" + normalizedHost + "_" + normalizedPort
}

// GenerateSmtpServerPasswordEnvVarName generates an environment variable name for an SMTP server's password
// based on its hostname and port. Format: SMTP_SERVER_PASSWORD_FOR_{NORMALIZED_HOST}_{NORMALIZED_PORT}
func GenerateSmtpServerPasswordEnvVarName(hostname, port string) string {
	normalizedHost := GenerateEnvVarName(hostname)
	normalizedPort := GenerateEnvVarName(port)
	return "SMTP_SERVER_PASSWORD_FOR_" + normalizedHost + "_" + normalizedPort
}

// SmtpServerCredentialsOverride represents the interface needed for SMTP server credential overrides
type SmtpServerCredentialsOverride interface {
	GetHost() string
	GetPort() string
	SetUsername(username string)
	SetPassword(password string)
}

// OverrideSmtpServerCredentialsFromEnv overrides SMTP server credentials from environment variables
// for any type that implements the SmtpServerCredentialsOverride interface
func OverrideSmtpServerCredentialsFromEnv(servers []SmtpServerCredentialsOverride) {
	for _, server := range servers {
		hostname := server.GetHost()
		port := server.GetPort()

		// Check for username override
		usernameEnvVar := GenerateSmtpServerUsernameEnvVarName(hostname, port)
		if username := os.Getenv(usernameEnvVar); username != "" {
			server.SetUsername(username)
		}

		// Check for password override
		passwordEnvVar := GenerateSmtpServerPasswordEnvVarName(hostname, port)
		if password := os.Getenv(passwordEnvVar); password != "" {
			server.SetPassword(password)
		}
	}
}
