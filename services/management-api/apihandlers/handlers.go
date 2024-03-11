package apihandlers

import (
	"net/http"
	"time"

	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	muDBConn           *muDB.ManagementUserDBService
	messagingDBConn    *messagingDB.MessagingDBService
	studyDBConn        *studyDB.StudyDBService
	tokenSignKey       string
	tokenExpiresIn     time.Duration
	allowedInstanceIDs []string
	filestorePath      string
}

func NewHTTPHandler(
	tokenSignKey string,
	tokenExpiresIn time.Duration,
	muDBConn *muDB.ManagementUserDBService,
	messagingDBConn *messagingDB.MessagingDBService,
	studyDBConn *studyDB.StudyDBService,
	allowedInstanceIDs []string,
	filestorePath string,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:       tokenSignKey,
		muDBConn:           muDBConn,
		messagingDBConn:    messagingDBConn,
		studyDBConn:        studyDBConn,
		allowedInstanceIDs: allowedInstanceIDs,
		tokenExpiresIn:     tokenExpiresIn,
		filestorePath:      filestorePath,
	}
}
