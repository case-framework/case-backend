package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/pkg/db"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"github.com/case-framework/case-backend/pkg/user-management/pwhash"
	"github.com/case-framework/case-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"

	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
)

// Environment variables
const (
	ENV_CONFIG_FILE_PATH = "CONFIG_FILE_PATH"

	ENV_GLOBAL_EMAIL_TEMPLATE_CONSTANTS_JSON = "GLOBAL_EMAIL_TEMPLATE_CONSTANTS_JSON"
	ENV_EMAIL_CLIENT_ADDRESS                 = "EMAIL_CLIENT_ADDRESS"
	ENV_EMAIL_CLIENT_API_KEY                 = "EMAIL_CLIENT_API_KEY"
	ENV_EMAIL_CLIENT_TIMEOUT                 = "EMAIL_CLIENT_TIMEOUT"
)

type ParticipantApiConfig struct {
	// Logging configs
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// Gin configs
	GinConfig struct {
		DebugMode    bool     `json:"debug_mode" yaml:"debug_mode"`
		AllowOrigins []string `json:"allow_origins" yaml:"allow_origins"`
		Port         string   `json:"port" yaml:"port"`

		// Mutual TLS configs
		MTLS struct {
			Use              bool                        `json:"use" yaml:"use"`
			CertificatePaths apihelpers.CertificatePaths `json:"certificate_paths" yaml:"certificate_paths"`
		} `json:"mtls" yaml:"mtls"`
	} `json:"gin_config" yaml:"gin_config"`

	// user management configs
	UserManagementConfig struct {
		PWHashing struct {
			Argon2Memory      uint32 `json:"argon2_memory" yaml:"argon2_memory"`
			Argon2Iterations  uint32 `json:"argon2_iterations" yaml:"argon2_iterations"`
			Argon2Parallelism uint8  `json:"argon2_parallelism" yaml:"argon2_parallelism"`
		} `json:"pw_hashing" yaml:"pw_hashing"`
		ParticipantUserJWTConfig struct {
			SignKey   string        `json:"sign_key" yaml:"sign_key"`
			ExpiresIn time.Duration `json:"expires_in" yaml:"expires_in"`
		} `json:"participant_user_jwt_config" yaml:"participant_user_jwt_config"`
		MaxNewUsersPer5Minutes           int            `json:"max_new_users_per_5_minutes" yaml:"max_new_users_per_5_minutes"`
		EmailContactVerificationTokenTTL time.Duration  `json:"email_contact_verification_token_ttl" yaml:"email_contact_verification_token_ttl"`
		WeekdayAssignationWeights        map[string]int `json:"weekday_assignation_weights" yaml:"weekday_assignation_weights"`
	} `json:"user_management_config" yaml:"user_management_config"`

	AllowedInstanceIDs []string `json:"allowed_instance_ids" yaml:"allowed_instance_ids"`

	// DB configs
	DBConfigs struct {
		StudyDB           db.DBConfigYaml `json:"study_db" yaml:"study_db"`
		ParticipantUserDB db.DBConfigYaml `json:"participant_user_db" yaml:"participant_user_db"`
		GlobalInfosDB     db.DBConfigYaml `json:"global_infos_db" yaml:"global_infos_db"`
		MessagingDB       db.DBConfigYaml `json:"messaging_db" yaml:"messaging_db"`
	} `json:"db_configs" yaml:"db_configs"`

	// Study module config
	StudyConfigs struct {
		GlobalSecret string `json:"global_secret" yaml:"global_secret"`
	} `json:"study_configs" yaml:"study_configs"`

	FilestorePath string `json:"filestore_path" yaml:"filestore_path"`

	MessagingConfigs messagingTypes.MessagingConfigs `json:"messaging_configs" yaml:"messaging_configs"`
}

func init() {
	// Read config from file
	yamlFile, err := os.ReadFile(os.Getenv(ENV_CONFIG_FILE_PATH))
	if err != nil {
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
	)

	if !conf.GinConfig.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// init argon2
	pwhash.InitArgonParams(
		conf.UserManagementConfig.PWHashing.Argon2Memory,
		conf.UserManagementConfig.PWHashing.Argon2Iterations,
		conf.UserManagementConfig.PWHashing.Argon2Parallelism,
	)

	umUtils.InitWeekdayAssignationStrategy(conf.UserManagementConfig.WeekdayAssignationWeights)

	initMessageSendingConfig()
	checkParticipantFilestorePath()
}

func checkParticipantFilestorePath() {
	// To store dynamically generated files
	fsPath := conf.FilestorePath
	if fsPath == "" {
		slog.Error("Filestore path not set - configure PARTICIPANT_FILESTORE_PATH env variable.")
		panic("Filestore path not set")
	}

	if _, err := os.Stat(fsPath); os.IsNotExist(err) {
		slog.Error("Filestore path does not exist", slog.String("path", fsPath))
		panic("Filestore path does not exist")
	}
}

func initMessageSendingConfig() {
	emailsending.InitMessageSendingVariables(
		loadEmailClientHTTPConfig(),
		conf.MessagingConfigs.GlobalEmailTemplateConstants,
	)
}

func loadEmailClientHTTPConfig() httpclient.ClientConfig {
	return httpclient.ClientConfig{
		RootURL: conf.MessagingConfigs.SmtpBridgeConfig.URL,
		APIKey:  conf.MessagingConfigs.SmtpBridgeConfig.APIKey,
		Timeout: conf.MessagingConfigs.SmtpBridgeConfig.RequestTimeout,
	}
}
