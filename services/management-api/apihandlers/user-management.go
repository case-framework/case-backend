package apihandlers

import (
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddUserManagementAPI(rg *gin.RouterGroup) {
	umGroup := rg.Group("/user-management")
	umGroup.Use(mw.GetAndValidateManagementUserJWT(h.tokenSignKey))
	umGroup.Use(mw.IsInstanceIDInJWTAllowed(h.allowedInstanceIDs))
	umGroup.Use(mw.IsAdminUser())
	{
		umGroup.GET("/management-users", h.getAllManagementUsers)
		umGroup.GET("/management-users/:userID", h.getManagementUser)
	}
}

func (h *HttpEndpoints) getAllManagementUsers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getManagementUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
