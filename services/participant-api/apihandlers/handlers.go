package apihandlers

import (
	"net/http"
	"time"

	globalinfosDB "github.com/case-framework/case-backend/pkg/db/global-infos"
	userDB "github.com/case-framework/case-backend/pkg/db/participant-user"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	studyDBConn        *studyDB.StudyDBService
	userDBConn         *userDB.ParticipantUserDBService
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
	studyDBConn *studyDB.StudyDBService,
	userDBConn *userDB.ParticipantUserDBService,
	globalInfosDBConn *globalinfosDB.GlobalInfosDBService,
	allowedInstanceIDs []string,
	globalStudySecret string,
	filestorePath string,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:       tokenSignKey,
		studyDBConn:        studyDBConn,
		userDBConn:         userDBConn,
		globalInfosDBConn:  globalInfosDBConn,
		allowedInstanceIDs: allowedInstanceIDs,
		globalStudySecret:  globalStudySecret,
		tokenExpiresIn:     tokenExpiresIn,
		filestorePath:      filestorePath,
	}
}
