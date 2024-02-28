package apihandlers

import (
	"net/http"
	"time"

	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	muDBConn           *muDB.ManagementUserDBService
	messagingDBConn    *messagingDB.MessagingDBService
	tokenSignKey       string
	tokenExpiresIn     time.Duration
	allowedInstanceIDs []string
}

func NewHTTPHandler(
	tokenSignKey string,
	tokenExpiresIn time.Duration,
	muDBConn *muDB.ManagementUserDBService,
	messagingDBConn *messagingDB.MessagingDBService,
	allowedInstanceIDs []string,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:       tokenSignKey,
		muDBConn:           muDBConn,
		messagingDBConn:    messagingDBConn,
		allowedInstanceIDs: allowedInstanceIDs,
		tokenExpiresIn:     tokenExpiresIn,
	}
}
