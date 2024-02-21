package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
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
}

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	conf = initConfig()
	if !conf.GinDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
}

func initConfig() Config {
	conf := Config{}
	conf.GinDebugMode = os.Getenv(ENV_GIN_DEBUG_MODE) == "true"
	conf.Port = os.Getenv(ENV_MANAGEMENT_API_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

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

	// Allowed instance IDs
	conf.AllowedInstanceIDs = strings.Split(os.Getenv(ENV_INSTANCE_IDS), ",")
	return conf
}

func readManagementUserDBConfig() db.DBConfig {
	connStr := os.Getenv(ENV_MANAGEMENT_USER_DB_CONNECTION_STR)
	username := os.Getenv(ENV_MANAGEMENT_USER_DB_USERNAME)
	password := os.Getenv(ENV_MANAGEMENT_USER_DB_PASSWORD)
	prefix := os.Getenv(ENV_MANAGEMENT_USER_DB_CONNECTION_PREFIX) // Used in test mode
	if connStr == "" || username == "" || password == "" {
		slog.Error("Couldn't read DB credentials for management user DB.")
		panic("Couldn't read DB credentials for management user DB.")
	}
	URI := fmt.Sprintf(`mongodb%s://%s:%s@%s`, prefix, username, password, connStr)

	var err error
	Timeout, err := strconv.Atoi(os.Getenv(ENV_MANAGEMENT_USER_DB_TIMEOUT))
	if err != nil {
		slog.Error("error during initConfig", slog.String("error", err.Error()), ENV_MANAGEMENT_USER_DB_TIMEOUT, os.Getenv(ENV_MANAGEMENT_USER_DB_TIMEOUT))
		panic(err)
	}

	IdleConnTimeout, err := strconv.Atoi(os.Getenv(ENV_MANAGEMENT_USER_DB_IDLE_CONN_TIMEOUT))
	if err != nil {
		slog.Error("error during initConfig", slog.String("error", err.Error()), ENV_MANAGEMENT_USER_DB_IDLE_CONN_TIMEOUT, os.Getenv(ENV_MANAGEMENT_USER_DB_IDLE_CONN_TIMEOUT))
		panic(err)
	}

	mps, err := strconv.Atoi(os.Getenv(ENV_MANAGEMENT_USER_DB_MAX_POOL_SIZE))
	MaxPoolSize := uint64(mps)
	if err != nil {
		slog.Error("error during initConfig", slog.String("error", err.Error()), ENV_MANAGEMENT_USER_DB_MAX_POOL_SIZE, os.Getenv(ENV_MANAGEMENT_USER_DB_MAX_POOL_SIZE))
		panic(err)
	}

	noCursorTimeout := os.Getenv(ENV_MANAGEMENT_USER_DB_USE_NO_CURSOR_TIMEOUT) == "true"
	DBNamePrefix := os.Getenv(ENV_MANAGEMENT_USER_DB_NAME_PREFIX)
	InstanceIDs := strings.Split(os.Getenv(ENV_INSTANCE_IDS), ",")

	return db.DBConfig{
		URI:             URI,
		Timeout:         Timeout,
		IdleConnTimeout: IdleConnTimeout,
		MaxPoolSize:     MaxPoolSize,
		NoCursorTimeout: noCursorTimeout,
		DBNamePrefix:    DBNamePrefix,
		InstanceIDs:     InstanceIDs,
	}
}
