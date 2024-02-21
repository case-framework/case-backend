package apihandlers

import (
	"net/http"

	muDB "github.com/case-framework/case-backend/pkg/db/management-user"
	"github.com/gin-gonic/gin"
)

func HealthCheckHandle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type HttpEndpoints struct {
	muDBConn           *muDB.ManagementUserDBService
	tokenSignKey       string
	allowedInstanceIDs []string
}

func NewHTTPHandler(
	tokenSignKey string,
	muDBConn *muDB.ManagementUserDBService,
	allowedInstanceIDs []string,
) *HttpEndpoints {
	return &HttpEndpoints{
		tokenSignKey:       tokenSignKey,
		muDBConn:           muDBConn,
		allowedInstanceIDs: allowedInstanceIDs,
	}
}
