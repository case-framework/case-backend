package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	"github.com/case-framework/case-backend/pkg/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	usermanagement "github.com/case-framework/case-backend/pkg/user-management"
	"github.com/case-framework/case-backend/pkg/utils"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"gopkg.in/yaml.v2"
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
	ENV_SMTP_BRIDGE_API_KEY          = "SMTP_BRIDGE_API_KEY"
	ENV_STUDY_GLOBAL_SECRET          = "STUDY_GLOBAL_SECRET"
)

type config struct {
	// Logging configs
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// DB configs
	DBConfigs struct {
		ParticipantUserDB db.DBConfigYaml `json:"participant_user_db" yaml:"participant_user_db"`
		GlobalInfosDB     db.DBConfigYaml `json:"global_infos_db" yaml:"global_infos_db"`
		MessagingDB       db.DBConfigYaml `json:"messaging_db" yaml:"messaging_db"`
		StudyDB           db.DBConfigYaml `json:"study_db" yaml:"study_db"`
	} `json:"db_configs" yaml:"db_configs"`

	InstanceIDs []string `json:"instance_ids" yaml:"instance_ids"`

	// user management configs
	UserManagementConfig struct {
		DeleteUnverifiedUsersAfter                 time.Duration `json:"delete_unverified_users_after" yaml:"delete_unverified_users_after"`
		SendReminderToConfirmAccountAfter          time.Duration `json:"send_reminder_to_confirm_account_after" yaml:"send_reminder_to_confirm_account_after"`
		EmailContactVerificationTokenTTL           time.Duration `json:"email_contact_verification_token_ttl" yaml:"email_contact_verification_token_ttl"`
		NotifyAfterInactiveFor                     time.Duration `json:"notify_after_inactive_for" yaml:"notify_after_inactive_for"`
		MarkForDeletionAfterInactivityNotification time.Duration `json:"mark_for_deletion_after_inactivity_notification" yaml:"mark_for_deletion_after_inactivity_notification"`
	} `json:"user_management_config" yaml:"user_management_config"`

	MessagingConfigs messagingTypes.MessagingConfigs `json:"messaging_configs" yaml:"messaging_configs"`

	// Study module config
	StudyConfigs struct {
		GlobalSecret string `json:"global_secret" yaml:"global_secret"`

		ExternalServices []studyengine.ExternalService `json:"external_services" yaml:"external_services"`
	} `json:"study_configs" yaml:"study_configs"`

	RunTasks struct {
		CleanUpUnverifiedUsers        bool `json:"clean_up_unverified_users" yaml:"clean_up_unverified_users"`
		SendReminderToConfirmAccounts bool `json:"send_reminder_to_confirm_accounts" yaml:"send_reminder_to_confirm_accounts"`
		HandleInactiveUsers           bool `json:"handle_inactive_users" yaml:"handle_inactive_users"`
		GenerateProfileIDLookup       bool `json:"generate_profile_id_lookup" yaml:"generate_profile_id_lookup"`
	} `json:"run_tasks" yaml:"run_tasks"`
}

var conf config

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

	// check config values:
	if conf.UserManagementConfig.DeleteUnverifiedUsersAfter == 0 {
		slog.Error("DeleteUnverifiedUsersAfter is not set")
		panic("DeleteUnverifiedUsersAfter is not set")
	}

	if conf.UserManagementConfig.SendReminderToConfirmAccountAfter == 0 {
		slog.Error("SendReminderToConfirmAccountAfter is not set")
		panic("SendReminderToConfirmAccountAfter is not set")
	}

	// init db
	initDBs()

	// init message sending
	initMessageSendingConfig()

	// init user management
	initUserManagement()

	// init study service
	initStudyService()
}

func secretsOverride() {
	// Override secrets from environment variables

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

	if apiKey := os.Getenv(ENV_SMTP_BRIDGE_API_KEY); apiKey != "" {
		conf.MessagingConfigs.SmtpBridgeConfig.APIKey = apiKey
	}

	if globalSecret := os.Getenv(ENV_STUDY_GLOBAL_SECRET); globalSecret != "" {
		conf.StudyConfigs.GlobalSecret = globalSecret
	}
	// Only check study global secret if tasks that use study service are enabled
	if (conf.RunTasks.CleanUpUnverifiedUsers || conf.RunTasks.HandleInactiveUsers) && conf.StudyConfigs.GlobalSecret == "" {
		slog.Error("Study global secret must not be empty, use the config file or the env variable STUDY_GLOBAL_SECRET")
		panic("Study global secret must not be empty")
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

func initDBs() {
	var err error
	participantUserDBService, err = userDB.NewParticipantUserDBService(db.DBConfigFromYamlObj(conf.DBConfigs.ParticipantUserDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Participant User DB", slog.String("error", err.Error()))
		panic(err)
	}

	globalInfosDBService, err = globalinfosDB.NewGlobalInfosDBService(db.DBConfigFromYamlObj(conf.DBConfigs.GlobalInfosDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Global Infos DB", slog.String("error", err.Error()))
		panic(err)
	}

	messagingDBService, err = messagingDB.NewMessagingDBService(db.DBConfigFromYamlObj(conf.DBConfigs.MessagingDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Messaging DB", slog.String("error", err.Error()))
		panic(err)
	}

	studyDBService, err = studyDB.NewStudyDBService(db.DBConfigFromYamlObj(conf.DBConfigs.StudyDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
		panic(err)
	}
}

func initMessageSendingConfig() {
	emailsending.InitMessageSendingVariables(
		nil, // no need for http client config, not sending emails directly
		conf.MessagingConfigs.GlobalEmailTemplateConstants,
		messagingDBService,
	)
}

func initUserManagement() {
	usermanagement.Init(participantUserDBService, globalInfosDBService)
}

func initStudyService() {
	study.Init(
		studyDBService,
		conf.StudyConfigs.GlobalSecret,
		conf.StudyConfigs.ExternalServices,
	)
}
