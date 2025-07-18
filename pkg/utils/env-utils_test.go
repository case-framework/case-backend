package utils

import "testing"

func TestGenerateEnvVarName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple alphanumeric name",
			input:    "myservice",
			expected: "MYSERVICE",
		},
		{
			name:     "name with hyphens",
			input:    "my-analytics-service",
			expected: "MY_ANALYTICS_SERVICE",
		},
		{
			name:     "name with spaces",
			input:    "my service name",
			expected: "MY_SERVICE_NAME",
		},
		{
			name:     "name with mixed characters",
			input:    "my-service_name.v2",
			expected: "MY_SERVICE_NAME_V2",
		},
		{
			name:     "name with leading/trailing special chars",
			input:    "-my_service-",
			expected: "MY_SERVICE",
		},
		{
			name:     "name already uppercase",
			input:    "MYSERVICE",
			expected: "MYSERVICE",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "---",
			expected: "",
		},
		{
			name:     "name with numbers",
			input:    "service-v1.2.3",
			expected: "SERVICE_V1_2_3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateEnvVarName(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateEnvVarName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateExternalServiceAPIKeyEnvVarName(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		expected    string
	}{
		{
			name:        "simple service name",
			serviceName: "analytics",
			expected:    "EXTERNAL_SERVICE_API_KEY_FOR_ANALYTICS",
		},
		{
			name:        "service name with hyphens",
			serviceName: "my-analytics-service",
			expected:    "EXTERNAL_SERVICE_API_KEY_FOR_MY_ANALYTICS_SERVICE",
		},
		{
			name:        "service name with dots and version",
			serviceName: "notification.hub.v2",
			expected:    "EXTERNAL_SERVICE_API_KEY_FOR_NOTIFICATION_HUB_V2",
		},
		{
			name:        "service name with spaces",
			serviceName: "my external service",
			expected:    "EXTERNAL_SERVICE_API_KEY_FOR_MY_EXTERNAL_SERVICE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateExternalServiceAPIKeyEnvVarName(tt.serviceName)
			if result != tt.expected {
				t.Errorf("GenerateExternalServiceAPIKeyEnvVarName(%q) = %q, want %q", tt.serviceName, result, tt.expected)
			}
		})
	}
}

func TestGenerateConfidentialResponseExportSecretEnvVarName(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		expected string
	}{
		{
			name:     "simple task name",
			taskName: "export1",
			expected: "CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_EXPORT1",
		},
		{
			name:     "task name with hyphens",
			taskName: "weekly-export",
			expected: "CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_WEEKLY_EXPORT",
		},
		{
			name:     "task name with mixed characters",
			taskName: "daily_export-v1.2",
			expected: "CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_DAILY_EXPORT_V1_2",
		},
		{
			name:     "task name with spaces",
			taskName: "my confidential export",
			expected: "CONF_RESP_EXPORT_STUDY_GLOBAL_SECRET_FOR_MY_CONFIDENTIAL_EXPORT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateConfidentialResponseExportSecretEnvVarName(tt.taskName)
			if result != tt.expected {
				t.Errorf("GenerateConfidentialResponseExportSecretEnvVarName(%q) = %q, want %q", tt.taskName, result, tt.expected)
			}
		})
	}
}
