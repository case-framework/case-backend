package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	httpclient "github.com/case-framework/case-backend/pkg/http-client"
	"github.com/case-framework/case-backend/pkg/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	"github.com/case-framework/case-backend/pkg/utils"
	"gopkg.in/yaml.v2"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	emailsending "github.com/case-framework/case-backend/pkg/messaging/email-sending"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
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
	ENV_STUDY_GLOBAL_SECRET = "STUDY_GLOBAL_SECRET"
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

	MessagingConfigs messagingTypes.MessagingConfigs `json:"messaging_configs" yaml:"messaging_configs"`

	RunTasks struct {
		ProcessOutgoingEmails     bool `json:"process_outgoing_emails" yaml:"process_outgoing_emails"`
		ScheduleHandler           bool `json:"schedule_handler" yaml:"schedule_handler"`
		StudyMessagesHandler      bool `json:"study_messages_handler" yaml:"study_messages_handler"`
		ResearcherMessagesHandler bool `json:"researcher_messages_handler" yaml:"researcher_messages_handler"`
	} `json:"run_tasks" yaml:"run_tasks"`

	Intervals struct {
		LastSendAttemptLockDuration time.Duration `json:"last_send_attempt_lock_duration" yaml:"last_send_attempt_lock_duration"`
		LoginTokenTTL               time.Duration `json:"login_token_ttl" yaml:"login_token_ttl"`
		UnsubscribeTokenTTL         time.Duration `json:"unsubscribe_token_ttl" yaml:"unsubscribe_token_ttl"`
	} `json:"intervals" yaml:"intervals"`

	// Study module config
	StudyConfigs struct {
		GlobalSecret string `json:"global_secret" yaml:"global_secret"`
	} `json:"study_configs" yaml:"study_configs"`
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

	// init db
	initDBs()

	// init message sending
	initMessageSendingConfig()

	// init study service
	if shouldInitStudyService() {
		initStudyService()
	}
}

func shouldInitStudyService() bool {
	return conf.RunTasks.ScheduleHandler || conf.RunTasks.StudyMessagesHandler || conf.RunTasks.ResearcherMessagesHandler
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
		loadEmailClientHTTPConfig(),
		conf.MessagingConfigs.GlobalEmailTemplateConstants,
		messagingDBService,
	)
}

func initStudyService() {
	study.Init(
		studyDBService,
		conf.StudyConfigs.GlobalSecret,
		[]studyengine.ExternalService{},
		nil,
	)
}

func loadEmailClientHTTPConfig() *httpclient.ClientConfig {
	return &httpclient.ClientConfig{
		RootURL: conf.MessagingConfigs.SmtpBridgeConfig.URL,
		APIKey:  conf.MessagingConfigs.SmtpBridgeConfig.APIKey,
		Timeout: conf.MessagingConfigs.SmtpBridgeConfig.RequestTimeout,
	}
}
