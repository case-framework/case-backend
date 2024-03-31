package apihandlers

import (
	"net/http"
	"time"

	studyDB "github.com/case-framework/case-backend/pkg/db/study"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	studyDBConn        *studyDB.StudyDBService
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
	allowedInstanceIDs []string,
	globalStudySecret string,
	filestorePath string,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:       tokenSignKey,
		studyDBConn:        studyDBConn,
		allowedInstanceIDs: allowedInstanceIDs,
		globalStudySecret:  globalStudySecret,
		tokenExpiresIn:     tokenExpiresIn,
		filestorePath:      filestorePath,
	}
}
