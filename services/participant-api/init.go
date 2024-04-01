package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/pkg/db"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	"github.com/case-framework/case-backend/pkg/user-management/pwhash"
	"github.com/case-framework/case-backend/pkg/utils"

	"github.com/gin-gonic/gin"

	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
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

	ENV_GLOBAL_INFOS_DB_CONNECTION_STR        = "GLOBAL_INFOS_DB_CONNECTION_STR"
	ENV_GLOBAL_INFOS_DB_USERNAME              = "GLOBAL_INFOS_DB_USERNAME"
	ENV_GLOBAL_INFOS_DB_PASSWORD              = "GLOBAL_INFOS_DB_PASSWORD"
	ENV_GLOBAL_INFOS_DB_CONNECTION_PREFIX     = "GLOBAL_INFOS_DB_CONNECTION_PREFIX"
	ENV_GLOBAL_INFOS_DB_NAME_PREFIX           = "GLOBAL_INFOS_DB_NAME_PREFIX"
	ENV_GLOBAL_INFOS_DB_TIMEOUT               = "GLOBAL_INFOS_DB_TIMEOUT"
	ENV_GLOBAL_INFOS_DB_IDLE_CONN_TIMEOUT     = "GLOBAL_INFOS_DB_IDLE_CONN_TIMEOUT"
	ENV_GLOBAL_INFOS_DB_USE_NO_CURSOR_TIMEOUT = "GLOBAL_INFOS_DB_USE_NO_CURSOR_TIMEOUT"
	ENV_GLOBAL_INFOS_DB_MAX_POOL_SIZE         = "GLOBAL_INFOS_DB_MAX_POOL_SIZE"

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

	ENV_MESSAGING_DB_CONNECTION_STR        = "MESSAGING_DB_CONNECTION_STR"
	ENV_MESSAGING_DB_USERNAME              = "MESSAGING_DB_USERNAME"
	ENV_MESSAGING_DB_PASSWORD              = "MESSAGING_DB_PASSWORD"
	ENV_MESSAGING_DB_CONNECTION_PREFIX     = "MESSAGING_DB_CONNECTION_PREFIX"
	ENV_MESSAGING_DB_NAME_PREFIX           = "MESSAGING_DB_NAME_PREFIX"
	ENV_MESSAGING_DB_TIMEOUT               = "MESSAGING_DB_TIMEOUT"
	ENV_MESSAGING_DB_IDLE_CONN_TIMEOUT     = "MESSAGING_DB_IDLE_CONN_TIMEOUT"
	ENV_MESSAGING_DB_USE_NO_CURSOR_TIMEOUT = "MESSAGING_DB_USE_NO_CURSOR_TIMEOUT"
	ENV_MESSAGING_DB_MAX_POOL_SIZE         = "MESSAGING_DB_MAX_POOL_SIZE"

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

	ENV_NEW_USER_RATE_LIMIT = "NEW_USER_RATE_LIMIT"

	ENV_WEEKDAY_ASSIGNATION_WEIGHTS = "WEEKDAY_ASSIGNATION_WEIGHTS"

	ENV_EMAIL_CONTACT_VERIFICATION_TOKEN_TTL = "EMAIL_CONTACT_VERIFICATION_TOKEN_TTL"

	ENV_GLOBAL_EMAIL_TEMPLATE_CONSTANTS_JSON = "GLOBAL_EMAIL_TEMPLATE_CONSTANTS_JSON"
	ENV_EMAIL_CLIENT_ADDRESS                 = "EMAIL_CLIENT_ADDRESS"
	ENV_EMAIL_CLIENT_API_KEY                 = "EMAIL_CLIENT_API_KEY"
	ENV_EMAIL_CLIENT_TIMEOUT                 = "EMAIL_CLIENT_TIMEOUT"
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

	StudyDBConfig           db.DBConfig `json:"study_db_config"`
	ParticipantUserDBConfig db.DBConfig `json:"participant_user_db_config"`
	GlobalInfosDBConfig     db.DBConfig `json:"global_infos_db_config"`
	MessagingDBConfig       db.DBConfig `json:"messaging_db_config"`

	StudyGlobalSecret string `json:"study_global_secret"`

	FilestorePath string `json:"filestore_path"`

	MaxNewUsersPer5Minutes int `json:"max_new_users_per_5_minutes"`

	EmailContactVerificationTokenTTL time.Duration `json:"email_contact_verification_token_ttl"`
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
	conf.ParticipantUserDBConfig = readParticipantUserDBConfig()
	conf.GlobalInfosDBConfig = readGlobalInfosDBConfig()
	conf.MessagingDBConfig = readMessagingDBConfig()

	// Study global secret
	conf.StudyGlobalSecret = os.Getenv(ENV_STUDY_GLOBAL_SECRET)
	if conf.StudyGlobalSecret == "" {
		slog.Error("Study global secret not set - configure STUDY_GLOBAL_SECRET env variable.")
		panic("Study global secret not set")
	}

	// Allowed instance IDs
	conf.AllowedInstanceIDs = readInstanceIDs()

	// Rate limit for new users
	conf.MaxNewUsersPer5Minutes = 50
	v := os.Getenv(ENV_NEW_USER_RATE_LIMIT)
	if v != "" {
		conf.MaxNewUsersPer5Minutes, err = strconv.Atoi(v)
		if err != nil {
			slog.Error("cannot parse max new users per 5 minutes", slog.String("value", v), slog.String("error", err.Error()))
			panic(err)
		}
	}

	// Weekday assignation weights override
	umUtils.InitWeekdayAssignationStrategyFromEnv(ENV_WEEKDAY_ASSIGNATION_WEIGHTS)

	// Email contact verification token TTL
	conf.EmailContactVerificationTokenTTL = 7 * 24 * time.Hour
	overrideVal := os.Getenv(ENV_EMAIL_CONTACT_VERIFICATION_TOKEN_TTL)
	if overrideVal != "" {
		conf.EmailContactVerificationTokenTTL, err = utils.ParseDurationString(overrideVal)
		if err != nil {
			slog.Error("couln't parse config value", slog.String("error", err.Error()), slog.String(ENV_EMAIL_CONTACT_VERIFICATION_TOKEN_TTL, overrideVal))
		}
	}
	slog.Debug("Email contact verification token TTL", slog.Float64("ttl", conf.EmailContactVerificationTokenTTL.Hours()))

	// Load message sending config
	initMessageSendingConfig()

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

func readParticipantUserDBConfig() db.DBConfig {
	return db.ReadDBConfigFromEnv(
		"participant user DB",
		ENV_PARTICIPANT_USER_DB_CONNECTION_STR,
		ENV_PARTICIPANT_USER_DB_USERNAME,
		ENV_PARTICIPANT_USER_DB_PASSWORD,
		ENV_PARTICIPANT_USER_DB_CONNECTION_PREFIX,
		ENV_PARTICIPANT_USER_DB_TIMEOUT,
		ENV_PARTICIPANT_USER_DB_IDLE_CONN_TIMEOUT,
		ENV_PARTICIPANT_USER_DB_MAX_POOL_SIZE,
		ENV_PARTICIPANT_USER_DB_USE_NO_CURSOR_TIMEOUT,
		ENV_PARTICIPANT_USER_DB_NAME_PREFIX,
		readInstanceIDs(),
	)
}

func readGlobalInfosDBConfig() db.DBConfig {
	return db.ReadDBConfigFromEnv(
		"global infos DB",
		ENV_GLOBAL_INFOS_DB_CONNECTION_STR,
		ENV_GLOBAL_INFOS_DB_USERNAME,
		ENV_GLOBAL_INFOS_DB_PASSWORD,
		ENV_GLOBAL_INFOS_DB_CONNECTION_PREFIX,
		ENV_GLOBAL_INFOS_DB_TIMEOUT,
		ENV_GLOBAL_INFOS_DB_IDLE_CONN_TIMEOUT,
		ENV_GLOBAL_INFOS_DB_MAX_POOL_SIZE,
		ENV_GLOBAL_INFOS_DB_USE_NO_CURSOR_TIMEOUT,
		ENV_GLOBAL_INFOS_DB_NAME_PREFIX,
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

func initMessageSendingConfig() {
	emailsending.InitMessageSendingVariables(
		loadEmailClientHTTPConfig(),
		loadGlobalEmailTemplateConstants(),
	)
}

func loadEmailClientHTTPConfig() httpclient.ClientConfig {
	timeOut := 60 * time.Second
	if v := os.Getenv(ENV_EMAIL_CLIENT_TIMEOUT); v != "" {
		var err error
		timeOut, err = time.ParseDuration(v)
		if err != nil {
			slog.Error("error parsing email client timeout", slog.String("value", v), slog.String("error", err.Error()))
		}
	}
	return httpclient.ClientConfig{
		RootURL: os.Getenv(ENV_EMAIL_CLIENT_ADDRESS),
		APIKey:  os.Getenv(ENV_EMAIL_CLIENT_API_KEY),
		Timeout: timeOut,
	}
}

func loadGlobalEmailTemplateConstants() map[string]string {
	// if filename defined through env variable, use it
	filename := os.Getenv(ENV_GLOBAL_EMAIL_TEMPLATE_CONSTANTS_JSON)
	if filename == "" {
		return nil
	}

	// load file
	file, err := os.Open(filename)
	if err != nil {
		slog.Error("error loading global email template constants file", slog.String("filename", filename), slog.String("error", err.Error()))
		return nil
	}
	defer file.Close()

	// parse file
	var config map[string]string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		slog.Error("error parsing global email template constants file", slog.String("filename", filename), slog.String("error", err.Error()))
		return nil
	}

	return config
}
