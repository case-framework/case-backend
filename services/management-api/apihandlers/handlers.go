package apihandlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	serviceInfos := make(map[string]interface{})
	infos, err := os.ReadFile("serviceInfos.json")
	if err != nil {
		slog.Debug("Error reading serviceInfos.json", slog.String("error", err.Error()))
	} else {
		err = json.Unmarshal(infos, &serviceInfos)
		if err != nil {
			slog.Debug("Error unmarshalling serviceInfos.json", slog.String("error", err.Error()))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"serviceInfos": serviceInfos,
	})
}

type HttpEndpoints struct {
	muDBConn           *muDB.ManagementUserDBService
	messagingDBConn    *messagingDB.MessagingDBService
	studyDBConn        *studyDB.StudyDBService
	participantUserDB  *userDB.ParticipantUserDBService
	globalInfosDBConn  *globalinfosDB.GlobalInfosDBService
	tokenSignKey       string
	tokenExpiresIn     time.Duration
	allowedInstanceIDs []string
	globalStudySecret  string
	filestorePath      string
}

func NewHTTPHandler(
	tokenSignKey string,
	tokenExpiresIn time.Duration,
	muDBConn *muDB.ManagementUserDBService,
	messagingDBConn *messagingDB.MessagingDBService,
	studyDBConn *studyDB.StudyDBService,
	participantUserDB *userDB.ParticipantUserDBService,
	globalInfosDBConn *globalinfosDB.GlobalInfosDBService,
	allowedInstanceIDs []string,
	globalStudySecret string,
	filestorePath string,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:       tokenSignKey,
		muDBConn:           muDBConn,
		messagingDBConn:    messagingDBConn,
		studyDBConn:        studyDBConn,
		participantUserDB:  participantUserDB,
		globalInfosDBConn:  globalInfosDBConn,
		allowedInstanceIDs: allowedInstanceIDs,
		globalStudySecret:  globalStudySecret,
		tokenExpiresIn:     tokenExpiresIn,
		filestorePath:      filestorePath,
	}
}
