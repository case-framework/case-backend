package main

import (
	"log/slog"
	"os"

	"github.com/case-framework/case-backend/pkg/db"
	"github.com/case-framework/case-backend/pkg/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	"github.com/case-framework/case-backend/pkg/utils"
	"gopkg.in/yaml.v2"

	studyDB "github.com/case-framework/case-backend/pkg/db/study"
)

// Environment variables
const (
	ENV_CONFIG_FILE_PATH = "CONFIG_FILE_PATH"

	// Variables to override "secrets" in the config file
	ENV_STUDY_DB_USERNAME = "STUDY_DB_USERNAME"
	ENV_STUDY_DB_PASSWORD = "STUDY_DB_PASSWORD"

	ENV_STUDY_GLOBAL_SECRET = "STUDY_GLOBAL_SECRET"
)

type config struct {
	// Logging configs
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// DB configs
	DBConfigs struct {
		StudyDB db.DBConfigYaml `json:"study_db" yaml:"study_db"`
	} `json:"db_configs" yaml:"db_configs"`

	InstanceIDs []string `json:"instance_ids" yaml:"instance_ids"`

	// Study module config
	StudyConfigs struct {
		GlobalSecret string `json:"global_secret" yaml:"global_secret"`

		ExternalServices []studyengine.ExternalService `json:"external_services" yaml:"external_services"`
	} `json:"study_configs" yaml:"study_configs"`

	CleanUpConfig struct {
		FilestorePath            string `json:"filestore_path" yaml:"filestore_path"`
		CleanOrphanedTaskResults bool   `json:"clean_orphaned_task_results" yaml:"clean_orphaned_task_results"`
	} `json:"clean_up_config" yaml:"clean_up_config"`
}

var conf config

var (
	studyDBService *studyDB.StudyDBService
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

	if studyGlobalSecret := os.Getenv(ENV_STUDY_GLOBAL_SECRET); studyGlobalSecret != "" {
		conf.StudyConfigs.GlobalSecret = studyGlobalSecret
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
	studyDBService, err = studyDB.NewStudyDBService(db.DBConfigFromYamlObj(conf.DBConfigs.StudyDB, conf.InstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
		panic(err)
	}
}

func initStudyService() {
	study.Init(
		studyDBService,
		conf.StudyConfigs.GlobalSecret,
		conf.StudyConfigs.ExternalServices,
	)
}
