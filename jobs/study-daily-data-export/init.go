package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/case-framework/case-backend/pkg/db"
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
)

type config struct {
	// Logging configs
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// DB configs
	DBConfigs struct {
		StudyDB db.DBConfigYaml `json:"study_db" yaml:"study_db"`
	} `json:"db_configs" yaml:"db_configs"`

	ResponseExports struct {
		RetentionDays int    `json:"retention_days" yaml:"retention_days"`
		ExportFormat  string `json:"export_format" yaml:"export_format"`
		Separator     string `json:"separator" yaml:"separator"`
		ShortKeys     bool   `json:"short_keys" yaml:"short_keys"`
		Sources       []struct {
			InstanceID string   `json:"instance_id" yaml:"instance_id"`
			StudyKey   string   `json:"study_key" yaml:"study_key"`
			SurveyKeys []string `json:"survey_keys" yaml:"survey_keys"`
		} `json:"sources" yaml:"sources"`
	} `json:"response_exports" yaml:"response_exports"`
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
	)

	// Override secrets from environment variables
	secretsOverride()

	// init db
	initDBs()

	if conf.ResponseExports.RetentionDays < 1 {
		err := fmt.Errorf("retention days must be greater than 0")
		slog.Error("Error reading config", slog.String("error", err.Error()))
		panic(err)
	}
}

func secretsOverride() {
	// Override secrets from environment variables

	if dbUsername := os.Getenv(ENV_STUDY_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.StudyDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_STUDY_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.StudyDB.Password = dbPassword
	}

}

func initDBs() {
	instanceIDs := getInstanceIDs()

	var err error
	studyDBService, err = studyDB.NewStudyDBService(db.DBConfigFromYamlObj(conf.DBConfigs.StudyDB, instanceIDs))
	if err != nil {
		slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
		panic(err)
	}
}

func getInstanceIDs() []string {
	instanceIDs := []string{}
	for _, source := range conf.ResponseExports.Sources {
		instanceIDs = append(instanceIDs, source.InstanceID)
	}
	return instanceIDs
}