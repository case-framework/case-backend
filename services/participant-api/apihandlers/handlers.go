package apihandlers

import (
	"net/http"
	"time"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type TTLs struct {
	AccessToken                   time.Duration
	EmailContactVerificationToken time.Duration
}

type HttpEndpoints struct {
	studyDBConn           *studyDB.StudyDBService
	userDBConn            *userDB.ParticipantUserDBService
	globalInfosDBConn     *globalinfosDB.GlobalInfosDBService
	messagingDBConn       *messagingDB.MessagingDBService
	tokenSignKey          string
	allowedInstanceIDs    []string
	globalStudySecret     string
	filestorePath         string
	maxNewUsersPer5Minute int
	ttls                  TTLs
}

func NewHTTPHandler(
	tokenSignKey string,
	studyDBConn *studyDB.StudyDBService,
	userDBConn *userDB.ParticipantUserDBService,
	globalInfosDBConn *globalinfosDB.GlobalInfosDBService,
	messagingDBConn *messagingDB.MessagingDBService,
	allowedInstanceIDs []string,
	globalStudySecret string,
	filestorePath string,
	maxNewUsersPer5Minute int,
	ttls TTLs,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:          tokenSignKey,
		studyDBConn:           studyDBConn,
		userDBConn:            userDBConn,
		globalInfosDBConn:     globalInfosDBConn,
		messagingDBConn:       messagingDBConn,
		allowedInstanceIDs:    allowedInstanceIDs,
		globalStudySecret:     globalStudySecret,
		filestorePath:         filestorePath,
		maxNewUsersPer5Minute: maxNewUsersPer5Minute,
		ttls:                  ttls,
	}
}
