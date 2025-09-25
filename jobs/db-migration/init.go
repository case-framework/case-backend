package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/case-framework/case-backend/pkg/db"
	"github.com/case-framework/case-backend/pkg/utils"
	"gopkg.in/yaml.v2"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	managementUserDB "github.com/case-framework/case-backend/pkg/db/management-user"
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
	ENV_MANAGEMENT_USER_DB_USERNAME  = "MANAGEMENT_USER_DB_USERNAME"
	ENV_MANAGEMENT_USER_DB_PASSWORD  = "MANAGEMENT_USER_DB_PASSWORD"
	ENV_GLOBAL_INFOS_DB_USERNAME     = "GLOBAL_INFOS_DB_USERNAME"
	ENV_GLOBAL_INFOS_DB_PASSWORD     = "GLOBAL_INFOS_DB_PASSWORD"
	ENV_MESSAGING_DB_USERNAME        = "MESSAGING_DB_USERNAME"
	ENV_MESSAGING_DB_PASSWORD        = "MESSAGING_DB_PASSWORD"
)

type config struct {
	// Logging configs
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// DB configs
	DBConfigs struct {
		GlobalInfosDB     db.DBConfigYaml `json:"global_infos_db" yaml:"global_infos_db"`
		ParticipantUserDB db.DBConfigYaml `json:"participant_user_db" yaml:"participant_user_db"`
		ManagementUserDB  db.DBConfigYaml `json:"management_user_db" yaml:"management_user_db"`
		MessagingDB       db.DBConfigYaml `json:"messaging_db" yaml:"messaging_db"`
		StudyDB           db.DBConfigYaml `json:"study_db" yaml:"study_db"`
	} `json:"db_configs" yaml:"db_configs"`

	InstanceIDs []string `json:"instance_ids" yaml:"instance_ids"`

	// Task configurations
	TaskConfigs TaskConfigs `json:"task_configs" yaml:"task_configs"`
}

// Explicit task configuration structs
type TaskConfigs struct {
	DropIndexes    DropIndexesConfig    `json:"drop_indexes" yaml:"drop_indexes"`
	CreateIndexes  CreateIndexesConfig  `json:"create_indexes" yaml:"create_indexes"`
	GetIndexes     GetIndexesConfig     `json:"get_indexes" yaml:"get_indexes"`
	MigrationTasks MigrationTasksConfig `json:"migration_tasks" yaml:"migration_tasks"`
}

type DropIndexesConfig struct {
	StudyDB           DropIndexesMode `json:"study_db" yaml:"study_db"`
	ParticipantUserDB DropIndexesMode `json:"participant_user_db" yaml:"participant_user_db"`
	ManagementUserDB  DropIndexesMode `json:"management_user_db" yaml:"management_user_db"`
	GlobalInfosDB     DropIndexesMode `json:"global_infos_db" yaml:"global_infos_db"`
	MessagingDB       DropIndexesMode `json:"messaging_db" yaml:"messaging_db"`
}

type CreateIndexesConfig struct {
	StudyDB           bool `json:"study_db" yaml:"study_db"`
	ParticipantUserDB bool `json:"participant_user_db" yaml:"participant_user_db"`
	ManagementUserDB  bool `json:"management_user_db" yaml:"management_user_db"`
	GlobalInfosDB     bool `json:"global_infos_db" yaml:"global_infos_db"`
	MessagingDB       bool `json:"messaging_db" yaml:"messaging_db"`
}

type GetIndexesConfig struct {
	StudyDB           string `json:"study_db" yaml:"study_db"`
	ParticipantUserDB string `json:"participant_user_db" yaml:"participant_user_db"`
	ManagementUserDB  string `json:"management_user_db" yaml:"management_user_db"`
	GlobalInfosDB     string `json:"global_infos_db" yaml:"global_infos_db"`
	MessagingDB       string `json:"messaging_db" yaml:"messaging_db"`
}

type MigrationTasksConfig struct {
	ParticipantUserContactInfosFix bool `json:"participant_user_contact_infos_fix" yaml:"participant_user_contact_infos_fix"`
}

type DropIndexesMode string

const (
	DropIndexesModeAll      DropIndexesMode = "all"
	DropIndexesModeDefaults DropIndexesMode = "defaults"
	DropIndexesModeNone     DropIndexesMode = "none"
)

func (mode DropIndexesMode) IsValid() bool {
	switch mode {
	case DropIndexesModeAll, DropIndexesModeDefaults, DropIndexesModeNone:
		return true
	default:
		return false
	}
}

func validateConfig() {
	validateDropIndexesMode("task_configs.drop_indexes.study_db", conf.TaskConfigs.DropIndexes.StudyDB)
	validateDropIndexesMode("task_configs.drop_indexes.participant_user_db", conf.TaskConfigs.DropIndexes.ParticipantUserDB)
	validateDropIndexesMode("task_configs.drop_indexes.management_user_db", conf.TaskConfigs.DropIndexes.ManagementUserDB)
	validateDropIndexesMode("task_configs.drop_indexes.global_infos_db", conf.TaskConfigs.DropIndexes.GlobalInfosDB)
	validateDropIndexesMode("task_configs.drop_indexes.messaging_db", conf.TaskConfigs.DropIndexes.MessagingDB)
}

func validateDropIndexesMode(field string, mode DropIndexesMode) {
	if !mode.IsValid() {
		panic(fmt.Sprintf("invalid drop indexes mode for %s: %q. Use one of: %v", field, mode, []DropIndexesMode{DropIndexesModeAll, DropIndexesModeDefaults, DropIndexesModeNone}))
	}
}

type RequiredDBs struct {
	StudyDB           bool
	ParticipantUserDB bool
	ManagementUserDB  bool
	GlobalInfosDB     bool
	MessagingDB       bool
}

var conf config

// Database service variables - initialized only for required databases based on task config
var (
	participantUserDBService *userDB.ParticipantUserDBService
	managementUserDBService  *managementUserDB.ManagementUserDBService
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

	validateConfig()

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

}

func secretsOverride() {
	// Override secrets from environment variables
	if dbUsername := os.Getenv(ENV_PARTICIPANT_USER_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.ParticipantUserDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_PARTICIPANT_USER_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.ParticipantUserDB.Password = dbPassword
	}

	if dbUsername := os.Getenv(ENV_MANAGEMENT_USER_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.ManagementUserDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_MANAGEMENT_USER_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.ManagementUserDB.Password = dbPassword
	}

	if dbUsername := os.Getenv(ENV_STUDY_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.StudyDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_STUDY_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.StudyDB.Password = dbPassword
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
}

type GetIndexesDBs struct {
	StudyDB           bool
	ParticipantUserDB bool
	ManagementUserDB  bool
	GlobalInfosDB     bool
	MessagingDB       bool
}

func shouldGetIndexesForDBs() GetIndexesDBs {
	getIndexes := conf.TaskConfigs.GetIndexes

	return GetIndexesDBs{
		StudyDB:           getIndexes.StudyDB != "" && getIndexes.StudyDB != "false",
		ParticipantUserDB: getIndexes.ParticipantUserDB != "" && getIndexes.ParticipantUserDB != "false",
		ManagementUserDB:  getIndexes.ManagementUserDB != "" && getIndexes.ManagementUserDB != "false",
		GlobalInfosDB:     getIndexes.GlobalInfosDB != "" && getIndexes.GlobalInfosDB != "false",
		MessagingDB:       getIndexes.MessagingDB != "" && getIndexes.MessagingDB != "false",
	}
}

// getRequiredDBs determines which databases need to be connected based on task configurations
func getRequiredDBs() RequiredDBs {
	requiredDBs := RequiredDBs{}

	dropIndexes := conf.TaskConfigs.DropIndexes
	createIndexes := conf.TaskConfigs.CreateIndexes
	migrationTasks := conf.TaskConfigs.MigrationTasks
	shouldGetIndexes := shouldGetIndexesForDBs()

	// Check drop_indexes configuration
	if dropIndexes.StudyDB != DropIndexesModeNone {
		requiredDBs.StudyDB = true
	}
	if dropIndexes.ParticipantUserDB != DropIndexesModeNone {
		requiredDBs.ParticipantUserDB = true
	}
	if dropIndexes.ManagementUserDB != DropIndexesModeNone {
		requiredDBs.ManagementUserDB = true
	}
	if dropIndexes.GlobalInfosDB != DropIndexesModeNone {
		requiredDBs.GlobalInfosDB = true
	}
	if dropIndexes.MessagingDB != DropIndexesModeNone {
		requiredDBs.MessagingDB = true
	}

	// Check create_indexes configuration
	if createIndexes.StudyDB {
		requiredDBs.StudyDB = true
	}
	if createIndexes.ParticipantUserDB {
		requiredDBs.ParticipantUserDB = true
	}
	if createIndexes.ManagementUserDB {
		requiredDBs.ManagementUserDB = true
	}
	if createIndexes.GlobalInfosDB {
		requiredDBs.GlobalInfosDB = true
	}
	if createIndexes.MessagingDB {
		requiredDBs.MessagingDB = true
	}

	// Check get_indexes configuration
	if shouldGetIndexes.StudyDB {
		requiredDBs.StudyDB = true
	}
	if shouldGetIndexes.ParticipantUserDB {
		requiredDBs.ParticipantUserDB = true
	}
	if shouldGetIndexes.ManagementUserDB {
		requiredDBs.ManagementUserDB = true
	}
	if shouldGetIndexes.GlobalInfosDB {
		requiredDBs.GlobalInfosDB = true
	}
	if shouldGetIndexes.MessagingDB {
		requiredDBs.MessagingDB = true
	}

	// Check migration_tasks configuration
	if migrationTasks.ParticipantUserContactInfosFix {
		requiredDBs.ParticipantUserDB = true
	}

	return requiredDBs
}

func initDBs() {
	// Get required databases based on task configurations
	requiredDBs := getRequiredDBs()

	var err error

	// Initialize only the required database services
	if requiredDBs.ParticipantUserDB {
		participantUserDBService, err = userDB.NewParticipantUserDBService(db.DBConfigFromYamlObj(conf.DBConfigs.ParticipantUserDB, conf.InstanceIDs))
		if err != nil {
			slog.Error("Error connecting to Participant User DB", slog.String("error", err.Error()))
			panic(err)
		}
	}

	if requiredDBs.ManagementUserDB {
		managementUserDBService, err = managementUserDB.NewManagementUserDBService(db.DBConfigFromYamlObj(conf.DBConfigs.ManagementUserDB, conf.InstanceIDs))
		if err != nil {
			slog.Error("Error connecting to Management User DB", slog.String("error", err.Error()))
			panic(err)
		}
	}

	if requiredDBs.GlobalInfosDB {
		globalInfosDBService, err = globalinfosDB.NewGlobalInfosDBService(db.DBConfigFromYamlObj(conf.DBConfigs.GlobalInfosDB, conf.InstanceIDs))
		if err != nil {
			slog.Error("Error connecting to Global Infos DB", slog.String("error", err.Error()))
			panic(err)
		}
	}

	if requiredDBs.MessagingDB {
		messagingDBService, err = messagingDB.NewMessagingDBService(db.DBConfigFromYamlObj(conf.DBConfigs.MessagingDB, conf.InstanceIDs))
		if err != nil {
			slog.Error("Error connecting to Messaging DB", slog.String("error", err.Error()))
			panic(err)
		}
	}

	if requiredDBs.StudyDB {
		studyDBService, err = studyDB.NewStudyDBService(db.DBConfigFromYamlObj(conf.DBConfigs.StudyDB, conf.InstanceIDs))
		if err != nil {
			slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
			panic(err)
		}
	}

	// Log which databases were connected
	slog.Info("Database connections established",
		slog.Bool("study_db", requiredDBs.StudyDB),
		slog.Bool("participant_user_db", requiredDBs.ParticipantUserDB),
		slog.Bool("management_user_db", requiredDBs.ManagementUserDB),
		slog.Bool("global_infos_db", requiredDBs.GlobalInfosDB),
		slog.Bool("messaging_db", requiredDBs.MessagingDB))
}
