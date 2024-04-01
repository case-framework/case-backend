package main

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/pkg/db"
	"github.com/case-framework/case-backend/pkg/user-management/pwhash"
	"github.com/case-framework/case-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Environment variables
const (
	ENV_GIN_DEBUG_MODE              = "GIN_DEBUG_MODE"
	ENV_PARTICIPANT_API_LISTEN_PORT = "PARTICIPANT_API_LISTEN_PORT"
	ENV_CORS_ALLOW_ORIGINS          = "CORS_ALLOW_ORIGINS"

	ENV_PARTICIPANT_USER_JWT_SIGN_KEY   = "PARTICIPANT_USER_JWT_SIGN_KEY"
	ENV_PARTICIPANT_USER_JWT_EXPIRES_IN = "PARTICIPANT_USER_JWT_EXPIRES_IN"

	ENV_REQUIRE_MUTUAL_TLS     = "REQUIRE_MUTUAL_TLS"
	ENV_MUTUAL_TLS_SERVER_CERT = "MUTUAL_TLS_SERVER_CERT"
	ENV_MUTUAL_TLS_SERVER_KEY  = "MUTUAL_TLS_SERVER_KEY"
	ENV_MUTUAL_TLS_CA_CERT     = "MUTUAL_TLS_CA_CERT"

	ENV_INSTANCE_IDS = "INSTANCE_IDS"

	ENV_PARTICIPANT_USER_DB_CONNECTION_STR        = "PARTICIPANT_USER_DB_CONNECTION_STR"
	ENV_PARTICIPANT_USER_DB_USERNAME              = "PARTICIPANT_USER_DB_USERNAME"
	ENV_PARTICIPANT_USER_DB_PASSWORD              = "PARTICIPANT_USER_DB_PASSWORD"
	ENV_PARTICIPANT_USER_DB_CONNECTION_PREFIX     = "PARTICIPANT_USER_DB_CONNECTION_PREFIX"
	ENV_PARTICIPANT_USER_DB_NAME_PREFIX           = "PARTICIPANT_USER_DB_NAME_PREFIX"
	ENV_PARTICIPANT_USER_DB_TIMEOUT               = "PARTICIPANT_USER_DB_TIMEOUT"
	ENV_PARTICIPANT_USER_DB_IDLE_CONN_TIMEOUT     = "PARTICIPANT_USER_DB_IDLE_CONN_TIMEOUT"
	ENV_PARTICIPANT_USER_DB_USE_NO_CURSOR_TIMEOUT = "PARTICIPANT_USER_DB_USE_NO_CURSOR_TIMEOUT"
	ENV_PARTICIPANT_USER_DB_MAX_POOL_SIZE         = "PARTICIPANT_USER_DB_MAX_POOL_SIZE"

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

	ENV_PARTICIPANT_FILESTORE_PATH = "PARTICIPANT_FILESTORE_PATH"

	ENV_ARGON2_MEMORY      = "ARGON2_MEMORY"
	ENV_ARGON2_ITERATIONS  = "ARGON2_ITERATIONS"
	ENV_ARGON2_PARALLELISM = "ARGON2_PARALLELISM"
)

type ParticipantApiConfig struct {
	// Gin configs
	GinDebugMode bool     `json:"gin_debug_mode"`
	AllowOrigins []string `json:"allow_origins"`
	Port         string   `json:"port"`

	// JWT configs
	ParticipantUserJWTSignKey   string        `json:"participant_user_jwt_sign_key"`
	ParticipantUserJWTExpiresIn time.Duration `json:"participant_user_jwt_expires_in"`

	AllowedInstanceIDs []string `json:"allowed_instance_ids"`

	// Mutual TLS configs
	UseMTLS          bool                        `json:"use_mtls"`
	CertificatePaths apihelpers.CertificatePaths `json:"certificate_paths"`

	StudyDBConfig db.DBConfig `json:"study_db_config"`

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

	pwhash.InitArgonParamsFromEnv(
		ENV_ARGON2_MEMORY,
		ENV_ARGON2_ITERATIONS,
		ENV_ARGON2_PARALLELISM,
	)
}

func initConfig() ParticipantApiConfig {
	conf := ParticipantApiConfig{}
	conf.GinDebugMode = os.Getenv(ENV_GIN_DEBUG_MODE) == "true"
	conf.Port = os.Getenv(ENV_PARTICIPANT_API_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

	conf.FilestorePath = getAndCheckParticipantFilestorePath()

	// JWT configs
	conf.ParticipantUserJWTSignKey = os.Getenv(ENV_PARTICIPANT_USER_JWT_SIGN_KEY)
	expInVal := os.Getenv(ENV_PARTICIPANT_USER_JWT_EXPIRES_IN)
	var err error
	conf.ParticipantUserJWTExpiresIn, err = utils.ParseDurationString(expInVal)
	if err != nil {
		slog.Error("error during initConfig", slog.String("error", err.Error()), ENV_PARTICIPANT_USER_JWT_EXPIRES_IN, expInVal)
		panic(err)
	}

	// Mutual TLS configs
	conf.UseMTLS = os.Getenv(ENV_REQUIRE_MUTUAL_TLS) == "true"
	conf.CertificatePaths = apihelpers.CertificatePaths{
		ServerCertPath: os.Getenv(ENV_MUTUAL_TLS_SERVER_CERT),
		ServerKeyPath:  os.Getenv(ENV_MUTUAL_TLS_SERVER_KEY),
		CACertPath:     os.Getenv(ENV_MUTUAL_TLS_CA_CERT),
	}

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

func getAndCheckParticipantFilestorePath() string {
	// To store dynamically generated files
	fsPath := os.Getenv(ENV_PARTICIPANT_FILESTORE_PATH)
	if fsPath == "" {
		slog.Error("Filestore path not set - configure PARTICIPANT_FILESTORE_PATH env variable.")
		panic("Filestore path not set")
	}

	if _, err := os.Stat(fsPath); os.IsNotExist(err) {
		slog.Error("Filestore path does not exist", slog.String("path", fsPath))
		panic("Filestore path does not exist")
	}
	return fsPath
}
