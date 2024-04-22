package types

import "time"

type MessagingConfigs struct {
	GlobalEmailTemplateConstants map[string]string `json:"global_email_template_constants" yaml:"global_email_template_constants"`

	SmtpBridgeConfig struct {
		URL            string        `json:"url" yaml:"url"`
		APIKey         string        `json:"api_key" yaml:"api_key"`
		RequestTimeout time.Duration `json:"request_timeout" yaml:"request_timeout"`
	} `json:"smtp_bridge_config" yaml:"smtp_bridge_config"`
}
