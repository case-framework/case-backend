package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	"github.com/case-framework/case-backend/pkg/db"
	"github.com/case-framework/case-backend/pkg/study"
	"github.com/case-framework/case-backend/pkg/study/studyengine"
	"github.com/case-framework/case-backend/pkg/utils"
	"gopkg.in/yaml.v2"

	"github.com/gin-gonic/gin"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
)

// Environment variables
const (
	ENV_CONFIG_FILE_PATH = "CONFIG_FILE_PATH"

	ENV_GIN_DEBUG_MODE             = "GIN_DEBUG_MODE"
	ENV_MANAGEMENT_API_LISTEN_PORT = "MANAGEMENT_API_LISTEN_PORT"
	ENV_CORS_ALLOW_ORIGINS         = "CORS_ALLOW_ORIGINS"

	ENV_MANAGEMENT_USER_JWT_SIGN_KEY   = "MANAGEMENT_USER_JWT_SIGN_KEY"
	ENV_MANAGEMENT_USER_JWT_EXPIRES_IN = "MANAGEMENT_USER_JWT_EXPIRES_IN"

	ENV_INSTANCE_IDS = "INSTANCE_IDS"

	ENV_MANAGEMENT_USER_DB_USERNAME  = "MANAGEMENT_USER_DB_USERNAME"
	ENV_MANAGEMENT_USER_DB_PASSWORD  = "MANAGEMENT_USER_DB_PASSWORD"
	ENV_PARTICIPANT_USER_DB_USERNAME = "PARTICIPANT_USER_DB_USERNAME"
	ENV_PARTICIPANT_USER_DB_PASSWORD = "PARTICIPANT_USER_DB_PASSWORD"
	ENV_GLOBAL_INFOS_DB_USERNAME     = "GLOBAL_INFOS_DB_USERNAME"
	ENV_GLOBAL_INFOS_DB_PASSWORD     = "GLOBAL_INFOS_DB_PASSWORD"
	ENV_MESSAGING_DB_USERNAME        = "MESSAGING_DB_USERNAME"
	ENV_MESSAGING_DB_PASSWORD        = "MESSAGING_DB_PASSWORD"
	ENV_STUDY_DB_USERNAME            = "STUDY_DB_USERNAME"
	ENV_STUDY_DB_PASSWORD            = "STUDY_DB_PASSWORD"

	ENV_STUDY_GLOBAL_SECRET = "STUDY_GLOBAL_SECRET"

	ENV_FILESTORE_PATH = "FILESTORE_PATH"
)

var (
	studyDBService           *studyDB.StudyDBService
	muDBService              *muDB.ManagementUserDBService
	messagingDBService       *messagingDB.MessagingDBService
	participantUserDBService *userDB.ParticipantUserDBService
	globalInfosDBService     *globalinfosDB.GlobalInfosDBService
)

type Config struct {
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

	// JWT configs
	ManagementUserJWTSignKey   string        `json:"management_user_jwt_sign_key" yaml:"management_user_jwt_sign_key"`
	ManagementUserJWTExpiresIn time.Duration `json:"management_user_jwt_expires_in" yaml:"management_user_jwt_expires_in"`

	AllowedInstanceIDs []string `json:"allowed_instance_ids" yaml:"allowed_instance_ids"`

	// DB configs
	DBConfigs struct {
		ParticipantUserDB db.DBConfigYaml `json:"participant_user_db" yaml:"participant_user_db"`
		ManagementUserDB  db.DBConfigYaml `json:"management_user_db" yaml:"management_user_db"`
		GlobalInfosDB     db.DBConfigYaml `json:"global_infos_db" yaml:"global_infos_db"`
		MessagingDB       db.DBConfigYaml `json:"messaging_db" yaml:"messaging_db"`
		StudyDB           db.DBConfigYaml `json:"study_db" yaml:"study_db"`
	} `json:"db_configs" yaml:"db_configs"`

	// Study module config
	StudyConfigs struct {
		GlobalSecret     string                        `json:"global_secret" yaml:"global_secret"`
		ExternalServices []studyengine.ExternalService `json:"external_services" yaml:"external_services"`
	} `json:"study_configs" yaml:"study_configs"`

	FilestorePath       string `json:"filestore_path" yaml:"filestore_path"`
	DailyFileExportPath string `json:"daily_file_export_path" yaml:"daily_file_export_path"`
}

func init() {
	conf = initConfig()
	if !conf.GinConfig.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Override secrets from environment variables
	secretsOverride()

	initDBs()

	initStudyService()
}

func initDBs() {
	var err error
	muDBService, err = muDB.NewManagementUserDBService(db.DBConfigFromYamlObj(conf.DBConfigs.ManagementUserDB, conf.AllowedInstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Management User DB", slog.String("error", err.Error()))
		panic(err)
	}

	messagingDBService, err = messagingDB.NewMessagingDBService(db.DBConfigFromYamlObj(conf.DBConfigs.MessagingDB, conf.AllowedInstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Messaging DB", slog.String("error", err.Error()))
		panic(err)
	}

	studyDBService, err = studyDB.NewStudyDBService(db.DBConfigFromYamlObj(conf.DBConfigs.StudyDB, conf.AllowedInstanceIDs))
	if err != nil {
		slog.Error("Error connecting to Study DB", slog.String("error", err.Error()))
		panic(err)
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
}

func initStudyService() {
	study.Init(
		studyDBService,
		conf.StudyConfigs.GlobalSecret,
		conf.StudyConfigs.ExternalServices,
	)
}

func getAndCheckFilestorePath() string {
	// To store dynamically generated files
	fsPath := os.Getenv(ENV_FILESTORE_PATH)
	if fsPath == "" {
		slog.Error("Filestore path not set")
		panic("Filestore path not set")
	}

	if _, err := os.Stat(fsPath); os.IsNotExist(err) {
		slog.Error("Filestore path does not exist", slog.String("path", fsPath))
		panic("Filestore path does not exist")
	}
	return fsPath
}

func initConfig() Config {
	conf := Config{}

	// Read config from file
	yamlFile, err := os.ReadFile(os.Getenv(ENV_CONFIG_FILE_PATH))
	if err != nil {
		fmt.Println("Error reading config file: " + err.Error())
		conf = Config{}
	}

	err = yaml.UnmarshalStrict(yamlFile, &conf)
	if err != nil {
		fmt.Println("Error reading config file: " + err.Error())
		conf = Config{}
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

	if os.Getenv(ENV_GIN_DEBUG_MODE) == "true" {
		conf.GinConfig.DebugMode = true
	}
	if port := os.Getenv(ENV_MANAGEMENT_API_LISTEN_PORT); port != "" {
		conf.GinConfig.Port = port
	}
	if origins := os.Getenv(ENV_CORS_ALLOW_ORIGINS); origins != "" {
		conf.GinConfig.AllowOrigins = strings.Split(origins, ",")
	}
	conf.FilestorePath = getAndCheckFilestorePath()

	// JWT configs
	conf.ManagementUserJWTSignKey = os.Getenv(ENV_MANAGEMENT_USER_JWT_SIGN_KEY)
	expInVal := os.Getenv(ENV_MANAGEMENT_USER_JWT_EXPIRES_IN)
	conf.ManagementUserJWTExpiresIn, err = utils.ParseDurationString(expInVal)
	if err != nil {
		slog.Error("error during initConfig", slog.String("error", err.Error()), ENV_MANAGEMENT_USER_JWT_EXPIRES_IN, expInVal)
		panic(err)
	}

	// Study global secret
	if studyGlobalSecret := os.Getenv(ENV_STUDY_GLOBAL_SECRET); studyGlobalSecret != "" {
		conf.StudyConfigs.GlobalSecret = studyGlobalSecret
	}
	if conf.StudyConfigs.GlobalSecret == "" {
		slog.Error("Study global secret not set - configure STUDY_GLOBAL_SECRET env variable.")
		panic("Study global secret not set")
	}

	// Allowed instance IDs
	envInstanceIDs := readInstanceIDs()
	if len(envInstanceIDs) > 0 {
		conf.AllowedInstanceIDs = envInstanceIDs
	}
	return conf
}

func readInstanceIDs() []string {
	instanceIDs := strings.Split(os.Getenv(ENV_INSTANCE_IDS), ",")
	// filter out empty strings
	var filteredInstanceIDs []string
	for _, instanceID := range instanceIDs {
		if instanceID != "" {
			filteredInstanceIDs = append(filteredInstanceIDs, instanceID)
		}
	}
	return filteredInstanceIDs
}

func secretsOverride() {
	// Override secrets from environment variables
	if dbUsername := os.Getenv(ENV_MANAGEMENT_USER_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.ManagementUserDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_MANAGEMENT_USER_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.ManagementUserDB.Password = dbPassword
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

	if dbUsername := os.Getenv(ENV_STUDY_DB_USERNAME); dbUsername != "" {
		conf.DBConfigs.StudyDB.Username = dbUsername
	}

	if dbPassword := os.Getenv(ENV_STUDY_DB_PASSWORD); dbPassword != "" {
		conf.DBConfigs.StudyDB.Password = dbPassword
	}

}
