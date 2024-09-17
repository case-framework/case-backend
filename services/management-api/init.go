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

	ENV_REQUIRE_MUTUAL_TLS     = "REQUIRE_MUTUAL_TLS"
	ENV_MUTUAL_TLS_SERVER_CERT = "MUTUAL_TLS_SERVER_CERT"
	ENV_MUTUAL_TLS_SERVER_KEY  = "MUTUAL_TLS_SERVER_KEY"
	ENV_MUTUAL_TLS_CA_CERT     = "MUTUAL_TLS_CA_CERT"

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

	ENV_LOG_TO_FILE     = "LOG_TO_FILE"
	ENV_LOG_FILENAME    = "LOG_FILENAME"
	ENV_LOG_MAX_SIZE    = "LOG_MAX_SIZE"
	ENV_LOG_MAX_AGE     = "LOG_MAX_AGE"
	ENV_LOG_MAX_BACKUPS = "LOG_MAX_BACKUPS"
	ENV_LOG_LEVEL       = "LOG_LEVEL"
	ENV_LOG_INCLUDE_SRC = "LOG_INCLUDE_SRC"

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

	FilestorePath string `json:"filestore_path"`
}

func init() {
	utils.ReadConfigFromEnvAndInitLogger(
		ENV_LOG_LEVEL,
		ENV_LOG_INCLUDE_SRC,
		ENV_LOG_TO_FILE,
		ENV_LOG_FILENAME,
		ENV_LOG_MAX_SIZE,
		ENV_LOG_MAX_AGE,
		ENV_LOG_MAX_BACKUPS,
	)

	conf = initConfig()
	if !conf.GinDebugMode {
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

	conf.GinDebugMode = os.Getenv(ENV_GIN_DEBUG_MODE) == "true"
	conf.Port = os.Getenv(ENV_MANAGEMENT_API_LISTEN_PORT)
	conf.AllowOrigins = strings.Split(os.Getenv(ENV_CORS_ALLOW_ORIGINS), ",")

	conf.FilestorePath = getAndCheckFilestorePath()

	// JWT configs
	conf.ManagementUserJWTSignKey = os.Getenv(ENV_MANAGEMENT_USER_JWT_SIGN_KEY)
	expInVal := os.Getenv(ENV_MANAGEMENT_USER_JWT_EXPIRES_IN)
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

	// Study global secret
	conf.StudyConfigs.GlobalSecret = os.Getenv(ENV_STUDY_GLOBAL_SECRET)
	if conf.StudyConfigs.GlobalSecret == "" {
		slog.Error("Study global secret not set - configure STUDY_GLOBAL_SECRET env variable.")
		panic("Study global secret not set")
	}

	// Allowed instance IDs
	conf.AllowedInstanceIDs = readInstanceIDs()
	return conf
}

func readInstanceIDs() []string {
	return strings.Split(os.Getenv(ENV_INSTANCE_IDS), ",")
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
