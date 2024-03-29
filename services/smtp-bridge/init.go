package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/case-framework/case-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Environment variables
const (
	ENV_GIN_DEBUG_MODE              = "GIN_DEBUG_MODE"
	ENV_SMTP_BRIDGE_API_LISTEN_PORT = "SMTP_BRIDGE_API_LISTEN_PORT"
	ENV_CORS_ALLOW_ORIGINS          = "CORS_ALLOW_ORIGINS"

	ENV_LOG_TO_FILE     = "LOG_TO_FILE"
	ENV_LOG_FILENAME    = "LOG_FILENAME"
	ENV_LOG_MAX_SIZE    = "LOG_MAX_SIZE"
	ENV_LOG_MAX_AGE     = "LOG_MAX_AGE"
	ENV_LOG_MAX_BACKUPS = "LOG_MAX_BACKUPS"
	ENV_LOG_LEVEL       = "LOG_LEVEL"
	ENV_LOG_INCLUDE_SRC = "LOG_INCLUDE_SRC"

	ENV_API_KEYS = "API_KEYS"
)

type config struct {
	// Gin configs
	GinDebugMode bool     `json:"gin_debug_mode"`
	AllowOrigins []string `json:"allow_origins"`
	Port         string   `json:"port"`

	ApiKeys []string `json:"api_keys"`
}

func init() {
	readConfigForAndInitLogger()

	conf = initConfig()
	if !conf.GinDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
}

func readConfigForAndInitLogger() {
	level := os.Getenv(ENV_LOG_LEVEL)
	includeSrc := os.Getenv(ENV_LOG_INCLUDE_SRC) == "true"
	logToFile := os.Getenv(ENV_LOG_TO_FILE) == "true"

	if !logToFile {
		utils.InitLogger(level, includeSrc, "", 0, 0, 0)
		return
	}

	logFilename := os.Getenv(ENV_LOG_FILENAME)
	logFileMaxSize, err := strconv.Atoi(os.Getenv(ENV_LOG_MAX_SIZE))
	if err != nil {
		panic(err)
	}
	logFileMaxAge, err := strconv.Atoi(os.Getenv(ENV_LOG_MAX_AGE))
	if err != nil {
		panic(err)
	}

	logFileMaxBackups, err := strconv.Atoi(os.Getenv(ENV_LOG_MAX_BACKUPS))
	if err != nil {
		panic(err)
	}

	utils.InitLogger(level, includeSrc, logFilename, logFileMaxSize, logFileMaxAge, logFileMaxBackups)
}

func initConfig() config {
	conf := config{}
	conf.GinDebugMode = os.Getenv(ENV_GIN_DEBUG_MODE) == "true"
	conf.Port = os.Getenv(ENV_SMTP_BRIDGE_API_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

	apiKeys := os.Getenv(ENV_API_KEYS)
	if apiKeys != "" {
		conf.ApiKeys = strings.Split(apiKeys, ",")
	}

	if len(conf.ApiKeys) == 0 {
		panic("No API keys provided for SMTP Bridge API. Please set the API_KEYS environment variable.")
	}

	return conf
}
