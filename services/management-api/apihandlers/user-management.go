package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	mUserDB "github.com/case-framework/case-backend/pkg/db/management-user"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddUserManagementAPI(rg *gin.RouterGroup) {
	umGroup := rg.Group("/user-management")
	umGroup.Use(mw.GetAndValidateManagementUserJWT(h.tokenSignKey))
	umGroup.Use(mw.IsInstanceIDInJWTAllowed(h.allowedInstanceIDs))
	{
		umGroup.GET("/management-users", h.getAllManagementUsers)

	}

	onlyAdminGroup := umGroup.Group("/")
	onlyAdminGroup.Use(mw.IsAdminUser())
	{
		umGroup.GET("/management-users/:userID", h.getManagementUser)
		umGroup.DELETE("/management-users/:userID", h.deleteManagementUser)
		umGroup.GET("/management-users/:userID/permissions", h.getManagementUserPermissions)
	}

}

func (h *HttpEndpoints) getAllManagementUsers(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("getAllManagementUsers: getting all users", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	users, err := h.muDBConn.GetAllUsers(token.InstanceID, token.IsAdmin)
	if err != nil {
		slog.Error("getAllManagementUsers: error retrieving users", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *HttpEndpoints) getManagementUser(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	slog.Info("getManagementUser: getting user", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	user, err := h.muDBConn.GetUserByID(token.InstanceID, userID)
	if err != nil {
		slog.Error("getManagementUser: error retrieving user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *HttpEndpoints) deleteManagementUser(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	if token.Subject == userID {
		slog.Error("deleteManagementUser: user cannot delete itself", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "user cannot delete itself"})
		return
	}

	slog.Info("deleteManagementUser: deleting user", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	// delete sessions
	err := h.muDBConn.DeleteSessionsByUserID(token.InstanceID, userID)
	if err != nil {
		slog.Error("deleteManagementUser: error deleting sessions", slog.String("error", err.Error()))
	}

	// delete permissions
	err = h.muDBConn.DeletePermissionsBySubject(token.InstanceID, userID, mUserDB.ManagementUserSubject)
	if err != nil {
		slog.Error("deleteManagementUser: error deleting permissions", slog.String("error", err.Error()))
	}

	// delete user
	err = h.muDBConn.DeleteUser(token.InstanceID, userID)
	if err != nil {
		slog.Error("deleteManagementUser: error deleting user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (h *HttpEndpoints) getManagementUserPermissions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	slog.Info("getManagementUserPermissions: getting user permissions", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	permissions, err := h.muDBConn.GetPermissionBySubject(token.InstanceID, userID, mUserDB.ManagementUserSubject)
	if err != nil {
		slog.Error("getManagementUserPermissions: error retrieving user permissions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}
