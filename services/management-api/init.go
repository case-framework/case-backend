package main

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/pkg/db"
	"github.com/case-framework/case-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Environment variables
const (
	ENV_GIN_DEBUG_MODE             = "GIN_DEBUG_MODE"
	ENV_MANAGEMENT_API_LISTEN_PORT = "MANAGEMENT_API_LISTEN_PORT"
	ENV_CORS_ALLOW_ORIGINS         = "CORS_ALLOW_ORIGINS"

	ENV_MANAGEMENT_USER_JWT_SIGN_KEY   = "MANAGEMENT_USER_JWT_SIGN_KEY"
	ENV_MANAGEMENT_USER_JWT_EXPIRES_IN = "MANAGEMENT_USER_JWT_EXPIRES_IN"

	ENV_REQUIRE_MUTUAL_TLS     = "REQUIRE_MUTUAL_TLS"
	ENV_MUTUAL_TLS_SERVER_CERT = "MUTUAL_TLS_SERVER_CERT"
	ENV_MUTUAL_TLS_SERVER_KEY  = "MUTUAL_TLS_SERVER_KEY"
	ENV_MUTUAL_TLS_CA_CERT     = "MUTUAL_TLS_CA_CERT"

	ENV_INSTANCE_IDS = "INSTANCE_IDS"

	ENV_MANAGEMENT_USER_DB_CONNECTION_STR        = "MANAGEMENT_USER_DB_CONNECTION_STR"
	ENV_MANAGEMENT_USER_DB_USERNAME              = "MANAGEMENT_USER_DB_USERNAME"
	ENV_MANAGEMENT_USER_DB_PASSWORD              = "MANAGEMENT_USER_DB_PASSWORD"
	ENV_MANAGEMENT_USER_DB_CONNECTION_PREFIX     = "MANAGEMENT_USER_DB_CONNECTION_PREFIX"
	ENV_MANAGEMENT_USER_DB_NAME_PREFIX           = "MANAGEMENT_USER_DB_NAME_PREFIX"
	ENV_MANAGEMENT_USER_DB_TIMEOUT               = "MANAGEMENT_USER_DB_TIMEOUT"
	ENV_MANAGEMENT_USER_DB_IDLE_CONN_TIMEOUT     = "MANAGEMENT_USER_DB_IDLE_CONN_TIMEOUT"
	ENV_MANAGEMENT_USER_DB_USE_NO_CURSOR_TIMEOUT = "MANAGEMENT_USER_DB_USE_NO_CURSOR_TIMEOUT"
	ENV_MANAGEMENT_USER_DB_MAX_POOL_SIZE         = "MANAGEMENT_USER_DB_MAX_POOL_SIZE"

	ENV_MESSAGING_DB_CONNECTION_STR        = "MESSAGING_DB_CONNECTION_STR"
	ENV_MESSAGING_DB_USERNAME              = "MESSAGING_DB_USERNAME"
	ENV_MESSAGING_DB_PASSWORD              = "MESSAGING_DB_PASSWORD"
	ENV_MESSAGING_DB_CONNECTION_PREFIX     = "MESSAGING_DB_CONNECTION_PREFIX"
	ENV_MESSAGING_DB_NAME_PREFIX           = "MESSAGING_DB_NAME_PREFIX"
	ENV_MESSAGING_DB_TIMEOUT               = "MESSAGING_DB_TIMEOUT"
	ENV_MESSAGING_DB_IDLE_CONN_TIMEOUT     = "MESSAGING_DB_IDLE_CONN_TIMEOUT"
	ENV_MESSAGING_DB_USE_NO_CURSOR_TIMEOUT = "MESSAGING_DB_USE_NO_CURSOR_TIMEOUT"
	ENV_MESSAGING_DB_MAX_POOL_SIZE         = "MESSAGING_DB_MAX_POOL_SIZE"

	ENV_STUDY_DB_CONNECTION_STR        = "STUDY_DB_CONNECTION_STR"
	ENV_STUDY_DB_USERNAME              = "STUDY_DB_USERNAME"
	ENV_STUDY_DB_PASSWORD              = "STUDY_DB_PASSWORD"
	ENV_STUDY_DB_CONNECTION_PREFIX     = "STUDY_DB_CONNECTION_PREFIX"
	ENV_STUDY_DB_NAME_PREFIX           = "STUDY_DB_NAME_PREFIX"
	ENV_STUDY_DB_TIMEOUT               = "STUDY_DB_TIMEOUT"
	ENV_STUDY_DB_IDLE_CONN_TIMEOUT     = "STUDY_DB_IDLE_CONN_TIMEOUT"
	ENV_STUDY_DB_USE_NO_CURSOR_TIMEOUT = "STUDY_DB_USE_NO_CURSOR_TIMEOUT"
	ENV_STUDY_DB_MAX_POOL_SIZE         = "STUDY_DB_MAX_POOL_SIZE"

	ENV_STUDY_GLOBAL_SECRET = "STUDY_GLOBAL_SECRET"

	ENV_LOG_TO_FILE     = "LOG_TO_FILE"
	ENV_LOG_FILENAME    = "LOG_FILENAME"
	ENV_LOG_MAX_SIZE    = "LOG_MAX_SIZE"
	ENV_LOG_MAX_AGE     = "LOG_MAX_AGE"
	ENV_LOG_MAX_BACKUPS = "LOG_MAX_BACKUPS"
	ENV_LOG_LEVEL       = "LOG_LEVEL"
	ENV_LOG_INCLUDE_SRC = "LOG_INCLUDE_SRC"

	ENV_FILESTORE_PATH = "FILESTORE_PATH"
)

type Config struct {
	// Gin configs
	GinDebugMode bool     `json:"gin_debug_mode"`
	AllowOrigins []string `json:"allow_origins"`
	Port         string   `json:"port"`

	// JWT configs
	ManagementUserJWTSignKey   string        `json:"management_user_jwt_sign_key"`
	ManagementUserJWTExpiresIn time.Duration `json:"management_user_jwt_expires_in"`

	AllowedInstanceIDs []string `json:"allowed_instance_ids"`

	// Mutual TLS configs
	UseMTLS          bool                        `json:"use_mtls"`
	CertificatePaths apihelpers.CertificatePaths `json:"certificate_paths"`

	ManagementUserDBConfig db.DBConfig `json:"management_user_db_config"`
	MessagingDBConfig      db.DBConfig `json:"messaging_db_config"`
	StudyDBConfig          db.DBConfig `json:"study_db_config"`

	StudyGlobalSecret string `json:"study_global_secret"`

	FilestorePath string `json:"filestore_path"`
}

func init() {
	utils.ReadConfigFromEnvAndInitLogger(
		ENV_LOG_LEVEL,
		ENV_LOG_INCLUDE_SRC,
		ENV_LOG_TO_FILE,
		ENV_LOG_FILENAME,
		ENV_LOG_MAX_SIZE,
		ENV_LOG_MAX_AGE,
		ENV_LOG_MAX_BACKUPS,
	)

	conf = initConfig()
	if !conf.GinDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
}

func getAndCheckFilestorePath() string {
	// To store dynamically generated files
	fsPath := os.Getenv(ENV_FILESTORE_PATH)
	if fsPath == "" {
		slog.Error("Filestore path not set")
		panic("Filestore path not set")
	}

	if _, err := os.Stat(fsPath); os.IsNotExist(err) {
		slog.Error("Filestore path does not exist", slog.String("path", fsPath))
		panic("Filestore path does not exist")
	}
	return fsPath
}

func initConfig() Config {
	conf := Config{}
	conf.GinDebugMode = os.Getenv(ENV_GIN_DEBUG_MODE) == "true"
	conf.Port = os.Getenv(ENV_MANAGEMENT_API_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

	conf.FilestorePath = getAndCheckFilestorePath()

	// JWT configs
	conf.ManagementUserJWTSignKey = os.Getenv(ENV_MANAGEMENT_USER_JWT_SIGN_KEY)
	expInVal := os.Getenv(ENV_MANAGEMENT_USER_JWT_EXPIRES_IN)
	var err error
	conf.ManagementUserJWTExpiresIn, err = utils.ParseDurationString(expInVal)
	if err != nil {
		slog.Error("error during initConfig", slog.String("error", err.Error()), ENV_MANAGEMENT_USER_JWT_EXPIRES_IN, expInVal)
		panic(err)
	}

	// Mutual TLS configs
	conf.UseMTLS = os.Getenv(ENV_REQUIRE_MUTUAL_TLS) == "true"
	conf.CertificatePaths = apihelpers.CertificatePaths{
		ServerCertPath: os.Getenv(ENV_MUTUAL_TLS_SERVER_CERT),
		ServerKeyPath:  os.Getenv(ENV_MUTUAL_TLS_SERVER_KEY),
		CACertPath:     os.Getenv(ENV_MUTUAL_TLS_CA_CERT),
	}

	// Management user db configs
	conf.ManagementUserDBConfig = readManagementUserDBConfig()

	// Messaging db configs
	conf.MessagingDBConfig = readMessagingDBConfig()

	// Study db configs
	conf.StudyDBConfig = readStudyDBConfig()

	// Study global secret
	conf.StudyGlobalSecret = os.Getenv(ENV_STUDY_GLOBAL_SECRET)
	if conf.StudyGlobalSecret == "" {
		slog.Error("Study global secret not set - configure STUDY_GLOBAL_SECRET env variable.")
		panic("Study global secret not set")
	}

	// Allowed instance IDs
	conf.AllowedInstanceIDs = readInstanceIDs()
	return conf
}

func readInstanceIDs() []string {
	return strings.Split(os.Getenv(ENV_INSTANCE_IDS), ",")
}

func readManagementUserDBConfig() db.DBConfig {
	return db.ReadDBConfigFromEnv(
		"management user DB",
		ENV_MANAGEMENT_USER_DB_CONNECTION_STR,
		ENV_MANAGEMENT_USER_DB_USERNAME,
		ENV_MANAGEMENT_USER_DB_PASSWORD,
		ENV_MANAGEMENT_USER_DB_CONNECTION_PREFIX,
		ENV_MANAGEMENT_USER_DB_TIMEOUT,
		ENV_MANAGEMENT_USER_DB_IDLE_CONN_TIMEOUT,
		ENV_MANAGEMENT_USER_DB_MAX_POOL_SIZE,
		ENV_MANAGEMENT_USER_DB_USE_NO_CURSOR_TIMEOUT,
		ENV_MANAGEMENT_USER_DB_NAME_PREFIX,
		readInstanceIDs(),
	)
}

func readMessagingDBConfig() db.DBConfig {
	return db.ReadDBConfigFromEnv(
		"messaging DB",
		ENV_MESSAGING_DB_CONNECTION_STR,
		ENV_MESSAGING_DB_USERNAME,
		ENV_MESSAGING_DB_PASSWORD,
		ENV_MESSAGING_DB_CONNECTION_PREFIX,
		ENV_MESSAGING_DB_TIMEOUT,
		ENV_MESSAGING_DB_IDLE_CONN_TIMEOUT,
		ENV_MESSAGING_DB_MAX_POOL_SIZE,
		ENV_MESSAGING_DB_USE_NO_CURSOR_TIMEOUT,
		ENV_MESSAGING_DB_NAME_PREFIX,
		readInstanceIDs(),
	)
}

func readStudyDBConfig() db.DBConfig {
	return db.ReadDBConfigFromEnv(
		"study DB",
		ENV_STUDY_DB_CONNECTION_STR,
		ENV_STUDY_DB_USERNAME,
		ENV_STUDY_DB_PASSWORD,
		ENV_STUDY_DB_CONNECTION_PREFIX,
		ENV_STUDY_DB_TIMEOUT,
		ENV_STUDY_DB_IDLE_CONN_TIMEOUT,
		ENV_STUDY_DB_MAX_POOL_SIZE,
		ENV_STUDY_DB_USE_NO_CURSOR_TIMEOUT,
		ENV_STUDY_DB_NAME_PREFIX,
		readInstanceIDs(),
	)
}
