package utils

import (
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
