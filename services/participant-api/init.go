package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	"github.com/case-framework/case-backend/pkg/db"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	"github.com/case-framework/case-backend/pkg/messaging/sms"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"github.com/case-framework/case-backend/pkg/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	studySender "github.com/case-framework/case-backend/pkg/study/studyengine/sender"
	usermanagement "github.com/case-framework/case-backend/pkg/user-management"
	"github.com/case-framework/case-backend/pkg/user-management/pwhash"
	"github.com/case-framework/case-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"

	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
)

// Environment variables
const (
	ENV_CONFIG_FILE_PATH = "CONFIG_FILE_PATH"

	// Variables to override "secrets" in the config file
	ENV_STUDY_DB_USERNAME            = "STUDY_DB_USERNAME"
	ENV_STUDY_DB_PASSWORD            = "STUDY_DB_PASSWORD"
	ENV_PARTICIPANT_USER_DB_USERNAME = "PARTICIPANT_USER_DB_USERNAME"
	ENV_PARTICIPANT_USER_DB_PASSWORD = "PARTICIPANT_USER_DB_PASSWORD"
	ENV_GLOBAL_INFOS_DB_USERNAME     = "GLOBAL_INFOS_DB_USERNAME"
	ENV_GLOBAL_INFOS_DB_PASSWORD     = "GLOBAL_INFOS_DB_PASSWORD"
	ENV_MESSAGING_DB_USERNAME        = "MESSAGING_DB_USERNAME"
	ENV_MESSAGING_DB_PASSWORD        = "MESSAGING_DB_PASSWORD"

	ENV_SMTP_BRIDGE_API_KEY = "SMTP_BRIDGE_API_KEY"
	ENV_SMS_GATEWAY_API_KEY = "SMS_GATEWAY_API_KEY"

	ENV_STUDY_GLOBAL_SECRET           = "STUDY_GLOBAL_SECRET"
	ENV_PARTICIPANT_USER_JWT_SIGN_KEY = "PARTICIPANT_USER_JWT_SIGN_KEY"
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
		OtpConfigs []middlewares.OTPConfig `json:"otp_configs" yaml:"otp_configs"`
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
		BlockedPasswordsFilePath         string         `json:"blocked_passwords_file_path" yaml:"blocked_passwords_file_path"`
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

		ExternalServices []studyengine.ExternalService `json:"external_services" yaml:"external_services"`
	} `json:"study_configs" yaml:"study_configs"`

	FilestorePath string `json:"filestore_path" yaml:"filestore_path"`

	MessagingConfigs messagingTypes.MessagingConfigs `json:"messaging_configs" yaml:"messaging_configs"`
}

var (
	participantUserDBService *userDB.ParticipantUserDBService
	globalInfosDBService     *globalinfosDB.GlobalInfosDBService
	messagingDBService       *messagingDB.MessagingDBService
	studyDBService           *studyDB.StudyDBService
)

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
		conf.Logging.CompressOldLogs,
		conf.Logging.IncludeBuildInfo,
	)

	// Override secrets from environment variables
	secretsOverride()

	// Init DBs
	initDBs()

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

	if conf.UserManagementConfig.BlockedPasswordsFilePath != "" {
		if err := umUtils.LoadBlockedPasswords(conf.UserManagementConfig.BlockedPasswordsFilePath); err != nil {
			panic(err)
		}
	}

	// init user management
	initUserManagement()

	// init message sending config
	initMessageSendingConfig()

	initStudyService()

	checkParticipantFilestorePath()
}

func secretsOverride() {
	// Override secrets from environment variables
	if apiKey := os.Getenv(ENV_SMTP_BRIDGE_API_KEY); apiKey != "" {
		conf.MessagingConfigs.SmtpBridgeConfig.APIKey = apiKey
	}

	if dbUsername := os.Getenv(ENV_STUDY_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.StudyDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_STUDY_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.StudyDB.Password = dbPassword
	}

	if dbUsername := os.Getenv(ENV_PARTICIPANT_USER_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.ParticipantUserDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_PARTICIPANT_USER_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.ParticipantUserDB.Password = dbPassword
	}

	if dbUsername := os.Getenv(ENV_GLOBAL_INFOS_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.GlobalInfosDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_GLOBAL_INFOS_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.GlobalInfosDB.Password = dbPassword
	}

	if dbUsername := os.Getenv(ENV_MESSAGING_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.MessagingDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_MESSAGING_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.MessagingDB.Password = dbPassword
	}

	if smsGatewayAPIKey := os.Getenv(ENV_SMS_GATEWAY_API_KEY); smsGatewayAPIKey != "" {
		if conf.MessagingConfigs.SMSConfig == nil {
			conf.MessagingConfigs.SMSConfig = &messagingTypes.SMSGatewayConfig{}
		}
		conf.MessagingConfigs.SMSConfig.APIKey = smsGatewayAPIKey
	}

	if studyGlobalSecret := os.Getenv(ENV_STUDY_GLOBAL_SECRET); studyGlobalSecret != "" {
		conf.StudyConfigs.GlobalSecret = studyGlobalSecret
	}

	if participantUserJwtSignKey := os.Getenv(ENV_PARTICIPANT_USER_JWT_SIGN_KEY); participantUserJwtSignKey != "" {
		conf.UserManagementConfig.ParticipantUserJWTConfig.SignKey = participantUserJwtSignKey
	}

	// Override API keys for external services
	for i := range conf.StudyConfigs.ExternalServices {
		service := &conf.StudyConfigs.ExternalServices[i]

		// Skip if name is not defined
		if service.Name == "" {
			continue
		}

		// Generate environment variable name from service name
		envVarName := utils.GenerateExternalServiceAPIKeyEnvVarName(service.Name)

		// Override if environment variable exists
		if apiKey := os.Getenv(envVarName); apiKey != "" {
			service.APIKey = apiKey
		}
	}
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

func initUserManagement() {
	usermanagement.Init(participantUserDBService, globalInfosDBService)
}

func initStudyService() {
	studyMessageSender := studySender.NewStudyMessageSender(
		studyDBService,
		participantUserDBService,
		messagingDBService,
		globalInfosDBService,
		studySender.MessageSenderConfig{
			LoginTokenTTL:                24 * time.Hour,
			GlobalEmailTemplateConstants: conf.MessagingConfigs.GlobalEmailTemplateConstants,
		},
	)

	study.Init(
		studyDBService,
		conf.StudyConfigs.GlobalSecret,
		conf.StudyConfigs.ExternalServices,
		studyMessageSender,
	)
}

func initMessageSendingConfig() {
	emailsending.InitMessageSendingVariables(
		loadEmailClientHTTPConfig(),
		conf.MessagingConfigs.GlobalEmailTemplateConstants,
		messagingDBService,
	)

	sms.Init(
		conf.MessagingConfigs.SMSConfig,
		messagingDBService,
	)
}

func loadEmailClientHTTPConfig() *httpclient.ClientConfig {
	return &httpclient.ClientConfig{
		RootURL: conf.MessagingConfigs.SmtpBridgeConfig.URL,
		APIKey:  conf.MessagingConfigs.SmtpBridgeConfig.APIKey,
		Timeout: conf.MessagingConfigs.SmtpBridgeConfig.RequestTimeout,
	}
}

func initDBs() {
	var err error
	studyDBService, err = studyDB.NewStudyDBService(db.DBConfigFromYamlObj(conf.DBConfigs.StudyDB, conf.AllowedInstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
		return
	}

	participantUserDBService, err = userDB.NewParticipantUserDBService(db.DBConfigFromYamlObj(conf.DBConfigs.ParticipantUserDB, conf.AllowedInstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Participant User DB", slog.String("error", err.Error()))
		return
	}

	globalInfosDBService, err = globalinfosDB.NewGlobalInfosDBService(db.DBConfigFromYamlObj(conf.DBConfigs.GlobalInfosDB, conf.AllowedInstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Global Infos DB", slog.String("error", err.Error()))
		return
	}

	messagingDBService, err = messagingDB.NewMessagingDBService(db.DBConfigFromYamlObj(conf.DBConfigs.MessagingDB, conf.AllowedInstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Messaging DB", slog.String("error", err.Error()))
		return
	}
}
