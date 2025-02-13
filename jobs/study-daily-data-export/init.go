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

type ResponseExportTask struct {
	InstanceID   string   `json:"instance_id" yaml:"instance_id"`
	StudyKey     string   `json:"study_key" yaml:"study_key"`
	SurveyKeys   []string `json:"survey_keys" yaml:"survey_keys"`
	ExtraCtxCols []string `json:"extra_context_columns" yaml:"extra_context_columns"`
	ExportFormat string   `json:"export_format" yaml:"export_format"`
	Separator    string   `json:"separator" yaml:"separator"`
	ShortKeys    bool     `json:"short_keys" yaml:"short_keys"`
}

type ConfidentialResponsesExportTask struct {
	InstanceID        string   `json:"instance_id" yaml:"instance_id"`
	StudyKey          string   `json:"study_key" yaml:"study_key"`
	StudyGlobalSecret string   `json:"study_global_secret" yaml:"study_global_secret"`
	Name              string   `json:"name" yaml:"name"`                       // optional name for the export file (used as "survey_key")
	RespKeyFilter     []string `json:"resp_key_filter" yaml:"resp_key_filter"` // optional filter for response keys to inlcude only these
	ExportFormat      string   `json:"export_format" yaml:"export_format"`     // csv or json
}

type config struct {
	// Logging configs
	Logging utils.LoggerConfig `json:"logging" yaml:"logging"`

	// DB configs
	DBConfigs struct {
		StudyDB db.DBConfigYaml `json:"study_db" yaml:"study_db"`
	} `json:"db_configs" yaml:"db_configs"`

	ExportPath string `json:"export_path" yaml:"export_path"`

	ResponseExports struct {
		RetentionDays int                  `json:"retention_days" yaml:"retention_days"`
		OverrideOld   bool                 `json:"override_old" yaml:"override_old"`
		ExportTasks   []ResponseExportTask `json:"export_tasks" yaml:"export_tasks"`
	} `json:"response_exports" yaml:"response_exports"`

	ConfidentialResponsesExports struct {
		PreservePreviousFiles bool                              `json:"preserve_previous_files" yaml:"preserve_previous_files"`
		ExportTasks           []ConfidentialResponsesExportTask `json:"export_tasks" yaml:"export_tasks"`
	} `json:"conf_resp_exports" yaml:"conf_resp_exports"`
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

	if conf.ResponseExports.RetentionDays < 1 {
		err := fmt.Errorf("retention days must be greater than 0")
		slog.Error("Error reading config", slog.String("error", err.Error()))
		panic(err)
	}

	if conf.ExportPath == "" {
		err := fmt.Errorf("export path must be set to define where to store the export files")
		slog.Error("Error reading config", slog.String("error", err.Error()))
		panic(err)
	}

	if _, err := os.Stat(conf.ExportPath); os.IsNotExist(err) {
		// create folder
		err = os.MkdirAll(conf.ExportPath, os.ModePerm)
		if err != nil {
			slog.Error("Error creating export path", slog.String("error", err.Error()))
			panic(err)
		}
		slog.Info("Created export path", slog.String("path", conf.ExportPath))
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
	for _, source := range conf.ResponseExports.ExportTasks {
		instanceIDs = append(instanceIDs, source.InstanceID)
	}
	return instanceIDs
}
