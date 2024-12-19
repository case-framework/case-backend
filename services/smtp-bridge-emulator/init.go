package main

import (
	"log/slog"
	"os"

	smtp_client "github.com/case-framework/case-backend/pkg/smtp-client"
	"github.com/case-framework/case-backend/pkg/utils"
	"gopkg.in/yaml.v2"
)

// Environment variables
const (
	ENV_CONFIG_FILE_PATH = "CONFIG_FILE_PATH"
)

type config struct {
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// Gin configs
	GinConfig struct {
		DebugMode    bool     `json:"debug_mode" yaml:"debug_mode"`
		AllowOrigins []string `json:"allow_origins" yaml:"allow_origins"`
		Port         string   `json:"port" yaml:"port"`
	} `json:"gin_config" yaml:"gin_config"`

	ApiKeys          []string `json:"api_keys" yaml:"api_keys"`
	SMTPServerConfig struct {
		HighPrio smtp_client.SmtpServerList `json:"high_prio" yaml:"high_prio"`
		LowPrio  smtp_client.SmtpServerList `json:"low_prio" yaml:"low_prio"`
	} `json:"smtp_server_config" yaml:"smtp_server_config"`
	EmailsDir string `yaml:"emails_dir"`
}

func init() {
	// Read config from file
	yamlFile, err := os.ReadFile(os.Getenv(ENV_CONFIG_FILE_PATH))
	if err != nil {
		slog.Error("Environment variable 'CONFIG_FILE_PATH' is not set correctly")
		panic(err)
	}

	err = yaml.UnmarshalStrict(yamlFile, &conf)
	if err != nil {
		panic(err)
	}

	// Init logger:
	utils.InitLogger(
		conf.Logging.LogLevel,
		conf.Logging.IncludeSrc,
		conf.Logging.LogToFile,
		conf.Logging.Filename,
		conf.Logging.MaxSize,
		conf.Logging.MaxAge,
		conf.Logging.MaxBackups,
		conf.Logging.CompressOldLogs,
		conf.Logging.IncludeBuildInfo,
	)

	if len(conf.ApiKeys) == 0 {
		panic("No API keys provided for SMTP Bridge API.")
	}

	if conf.EmailsDir == "" {
		panic("Emails directory to store emails not provided for SMTP Bridge Emulator API.")
	}
}
