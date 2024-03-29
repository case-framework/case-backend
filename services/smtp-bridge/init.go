package main

import (
	"os"
	"strconv"

	"github.com/case-framework/case-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type Config struct {
	// Gin configs
	GinDebugMode bool     `json:"gin_debug_mode"`
	AllowOrigins []string `json:"allow_origins"`
	Port         string   `json:"port"`

	ApiKeys []string `json:"api_keys"`
}

func init() {
	initLogger()

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
