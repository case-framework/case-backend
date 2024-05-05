package apihandlers

import (
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddUserManagementAPI(rg *gin.RouterGroup) {
	userGroup := rg.Group("/user")
	userGroup.Use(mw.GetAndValidateParticipantUserJWT(h.tokenSignKey))
	{
		userGroup.GET("/", h.getUser)
	}
}

func (h *HttpEndpoints) getUser(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	user, err := h.userDBConn.GetUser(token.InstanceID, token.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get user"})
		return
	}
	user.Account.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}
