package apihandlers

import (
	"net/http"
	"time"

	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	muDBConn           *muDB.ManagementUserDBService
	tokenSignKey       string
	tokenExpiresIn     time.Duration
	allowedInstanceIDs []string
}

func NewHTTPHandler(
	tokenSignKey string,
	tokenExpiresIn time.Duration,
	muDBConn *muDB.ManagementUserDBService,
	allowedInstanceIDs []string,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:       tokenSignKey,
		muDBConn:           muDBConn,
		allowedInstanceIDs: allowedInstanceIDs,
		tokenExpiresIn:     tokenExpiresIn,
	}
}
