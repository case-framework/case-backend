package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/case-framework/case-backend/pkg/utils"
	"gopkg.in/yaml.v2"
)

// Environment variables
const (
	ENV_CONFIG_FILE_PATH = "CONFIG_FILE_PATH"

	// Variables to override "secrets" in the config file
	ENV_API_KEYS = "API_KEYS"
)

type config struct {
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// Gin configs
	GinConfig struct {
		DebugMode    bool     `json:"debug_mode" yaml:"debug_mode"`
		AllowOrigins []string `json:"allow_origins" yaml:"allow_origins"`
		Port         string   `json:"port" yaml:"port"`
	} `json:"gin_config" yaml:"gin_config"`

	ApiKeys   []string `json:"api_keys" yaml:"api_keys"`
	EmailsDir string   `yaml:"emails_dir"`
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

	overrideFromEnv()

	if len(conf.ApiKeys) == 0 {
		panic("No API keys provided for SMTP Bridge API.")
	}

	if conf.EmailsDir == "" {
		panic("Emails directory to store emails not provided for SMTP Bridge Emulator API.")
	}
}

func overrideFromEnv() {
	// Override secrets from environment variables
	if apiKeys := os.Getenv(ENV_API_KEYS); apiKeys != "" {
		conf.ApiKeys = []string{}
		for _, apiKey := range strings.Split(apiKeys, ",") {
			key := strings.TrimSpace(apiKey)
			if key != "" {
				conf.ApiKeys = append(conf.ApiKeys, key)
			}
		}
	}
}
