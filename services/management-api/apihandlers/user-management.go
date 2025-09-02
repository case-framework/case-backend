package apihandlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	mUserDB "github.com/case-framework/case-backend/pkg/db/management-user"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	pc "github.com/case-framework/case-backend/pkg/permission-checker"
	"github.com/case-framework/case-backend/pkg/user-management/utils"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"

	studyService "github.com/case-framework/case-backend/pkg/study"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddUserManagementAPI(rg *gin.RouterGroup) {
	umGroup := rg.Group("/user-management")
	umGroup.Use(mw.ManagementAuthMiddleware(h.tokenSignKey, h.allowedInstanceIDs, h.muDBConn))

	// all management users can see other users (though not all details if not admin)
	{
		umGroup.GET("/management-users", h.getAllManagementUsers)
	}

	managementUsersGroup := umGroup.Group("/management-users")
	managementUsersGroup.Use(mw.IsAdminUser())
	{
		managementUsersGroup.GET("/:userID", h.getManagementUser)
		managementUsersGroup.DELETE("/:userID", h.deleteManagementUser)
		managementUsersGroup.GET("/:userID/permissions", h.getManagementUserPermissions)
		managementUsersGroup.POST("/:userID/permissions", mw.RequirePayload(), h.createManagementUserPermission)
		managementUsersGroup.DELETE("/:userID/permissions/:permissionID", h.deleteManagementUserPermission)
		managementUsersGroup.PUT("/:userID/permissions/:permissionID/limiter", mw.RequirePayload(), h.updateManagementUserPermissionLimiter)

		// App roles
		managementUsersGroup.GET("/:userID/app-roles", h.getManagementUserAppRoles)
		managementUsersGroup.POST("/:userID/app-roles/:appRoleTemplateID", h.addManagementUserAppRole)
		managementUsersGroup.DELETE("/:userID/app-roles/:appRoleID", h.deleteManagementUserAppRole)
	}

	// App role management
	appRolesGroup := managementUsersGroup.Group("/app-roles")
	appRolesGroup.Use(mw.IsAdminUser())
	{
		appRoleTemplatesGroup := appRolesGroup.Group("/templates")
		{
			appRoleTemplatesGroup.GET("/", h.getAllAppRoleTemplates)
			appRoleTemplatesGroup.POST("/", mw.RequirePayload(), h.createAppRoleTemplate)
			appRoleTemplatesGroup.GET("/:appRoleTemplateID", h.getAppRoleTemplateByID)
			appRoleTemplatesGroup.PUT("/:appRoleTemplateID", mw.RequirePayload(), h.updateAppRoleTemplate)
			appRoleTemplatesGroup.DELETE("/delete/:appRoleTemplateID", h.deleteAppRoleTemplate)
			appRoleTemplatesGroup.DELETE("/delete-for-app/:appName", h.deleteAppRoleTemplateForApp)
		}

		appRolesGroup.GET("/", h.getAllAppRoles)
		appRolesGroup.DELETE("/delete/:appName", h.deleteAllAppRolesForApp)
	}

	participantUsersGroup := umGroup.Group("/participant-users")
	{
		participantUsersGroup.POST("/request-deletion", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType: pc.RESOURCE_TYPE_USERS,
				ResourceKeys: []string{pc.RESOURCE_KEY_STUDY_ALL},
				Action:       pc.ACTION_DELETE_USERS,
			},
			nil,
			h.requestParticipantUserDeletion,
		))
	}

	serviceAccountsGroup := umGroup.Group("/service-accounts")
	serviceAccountsGroup.Use(mw.IsAdminUser())
	{
		serviceAccountsGroup.GET("/", h.getAllServiceAccounts)
		serviceAccountsGroup.POST("/", mw.RequirePayload(), h.createServiceAccount)
		serviceAccountsGroup.GET("/:serviceAccountID", h.getServiceAccount)
		serviceAccountsGroup.PUT("/:serviceAccountID", mw.RequirePayload(), h.updateServiceAccount)
		serviceAccountsGroup.GET("/:serviceAccountID/api-keys", h.getServiceAccountAPIKeys)
		serviceAccountsGroup.POST("/:serviceAccountID/api-keys", mw.RequirePayload(), h.createServiceAccountAPIKey)
		serviceAccountsGroup.DELETE("/:serviceAccountID/api-keys/:apiKeyID", h.deleteServiceAccountAPIKey)
		serviceAccountsGroup.DELETE("/:serviceAccountID", h.deleteServiceAccount)
		serviceAccountsGroup.GET("/:serviceAccountID/permissions", h.getServiceAccountPermissions)
		serviceAccountsGroup.POST("/:serviceAccountID/permissions", mw.RequirePayload(), h.createServiceAccountPermission)
		serviceAccountsGroup.DELETE("/:serviceAccountID/permissions/:permissionID", h.deleteServiceAccountPermission)
		serviceAccountsGroup.PUT("/:serviceAccountID/permissions/:permissionID/limiter", mw.RequirePayload(), h.updateServiceAccountPermissionLimiter)
	}

}

func (h *HttpEndpoints) getAllManagementUsers(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("getting all users", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	users, err := h.muDBConn.GetAllUsers(token.InstanceID, token.IsAdmin)
	if err != nil {
		slog.Error("error retrieving users", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *HttpEndpoints) getManagementUser(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	slog.Info("getting user", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	user, err := h.muDBConn.GetUserByID(token.InstanceID, userID)
	if err != nil {
		slog.Error("error retrieving user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *HttpEndpoints) deleteManagementUser(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	if token.Subject == userID {
		slog.Error("user cannot delete itself", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "user cannot delete itself"})
		return
	}

	slog.Info("deleting user", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	// delete sessions
	err := h.muDBConn.DeleteSessionsByUserID(token.InstanceID, userID)
	if err != nil {
		slog.Error("error deleting sessions", slog.String("error", err.Error()))
	}

	// delete permissions
	err = h.muDBConn.DeletePermissionsBySubject(token.InstanceID, userID, pc.SUBJECT_TYPE_MANAGEMENT_USER)
	if err != nil {
		slog.Error("error deleting permissions", slog.String("error", err.Error()))
	}

	// delete user
	err = h.muDBConn.DeleteUser(token.InstanceID, userID)
	if err != nil {
		slog.Error("error deleting user", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (h *HttpEndpoints) getManagementUserPermissions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	slog.Info("getting user permissions", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	permissions, err := h.muDBConn.GetPermissionBySubject(token.InstanceID, userID, pc.SUBJECT_TYPE_MANAGEMENT_USER)
	if err != nil {
		slog.Error("error retrieving user permissions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

func (h *HttpEndpoints) createManagementUserPermission(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	var newPerm mUserDB.Permission
	if err := c.ShouldBindJSON(&newPerm); err != nil {
		slog.Error("error binding permission", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing payload"})
		return
	}

	slog.Info("creating user permission", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	_, err := h.muDBConn.GetUserByID(token.InstanceID, userID)
	if err != nil {
		slog.Error("user not found", slog.String("userID", userID), slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	newPerm.SubjectType = pc.SUBJECT_TYPE_MANAGEMENT_USER
	newPerm.SubjectID = userID

	permission, err := h.muDBConn.CreatePermission(
		token.InstanceID,
		newPerm.SubjectID,
		newPerm.SubjectType,
		newPerm.ResourceType,
		newPerm.ResourceKey,
		newPerm.Action,
		newPerm.Limiter,
	)
	if err != nil {
		slog.Error("error creating user permission", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating user permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permission": permission})
}

func (h *HttpEndpoints) deleteManagementUserPermission(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")
	permissionID := c.Param("permissionID")

	slog.Info("deleting user permission", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("permissionForUser", userID), slog.String("permissionID", permissionID))

	err := h.muDBConn.DeletePermission(token.InstanceID, permissionID)
	if err != nil {
		slog.Error("error deleting user permission", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting user permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission deleted"})
}

func (h *HttpEndpoints) updateManagementUserPermissionLimiter(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")
	permissionID := c.Param("permissionID")

	var newLimiter mUserDB.Permission
	if err := c.ShouldBindJSON(&newLimiter); err != nil {
		slog.Error("error binding permission", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing payload"})
		return
	}

	slog.Info("updating user permission limiter", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("permissionForUser", userID), slog.String("permissionID", permissionID))

	err := h.muDBConn.UpdatePermissionLimiter(token.InstanceID, permissionID, newLimiter.Limiter)
	if err != nil {
		slog.Error("error updating user permission limiter", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating user permission limiter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission limiter updated"})
}

func (h *HttpEndpoints) getManagementUserAppRoles(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")

	slog.Info("getting user app roles", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID))

	appRoles, err := h.muDBConn.GetAppRolesForSubject(token.InstanceID, userID)
	if err != nil {
		slog.Error("error retrieving user app roles", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting user app roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appRoles": appRoles})
}

func permissionExists(existingPermissions []*mUserDB.Permission, permission *mUserDB.Permission) bool {
	for _, existingPermission := range existingPermissions {
		if existingPermission.ResourceType == permission.ResourceType && existingPermission.ResourceKey == permission.ResourceKey && existingPermission.Action == permission.Action {
			return true
		}
	}
	return false
}

func (h *HttpEndpoints) addManagementUserAppRole(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")
	appRoleTemplateID := c.Param("appRoleTemplateID")

	slog.Info("adding user app role", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID), slog.String("appRoleTemplateID", appRoleTemplateID))

	// find template for app role
	appRoleTemplate, err := h.muDBConn.GetAppRoleTemplateByID(token.InstanceID, appRoleTemplateID)
	if err != nil {
		slog.Error("error retrieving app role template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error retrieving app role template"})
		return
	}

	err = h.muDBConn.AddAppRoleForSubject(token.InstanceID, userID, pc.SUBJECT_TYPE_MANAGEMENT_USER, appRoleTemplate.AppName, appRoleTemplate.Role)
	if err != nil {
		slog.Error("error adding user app role", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error adding user app role"})
		return
	}

	userPermissions, err := h.muDBConn.GetPermissionBySubject(token.InstanceID, userID, pc.SUBJECT_TYPE_MANAGEMENT_USER)
	if err != nil {
		slog.Error("error retrieving user permissions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error retrieving user permissions"})
		return
	}

	for _, requiredPerm := range appRoleTemplate.RequiredPermissions {
		if !permissionExists(userPermissions, &requiredPerm) {
			_, err := h.muDBConn.CreatePermission(token.InstanceID, userID, pc.SUBJECT_TYPE_MANAGEMENT_USER, requiredPerm.ResourceType, requiredPerm.ResourceKey, requiredPerm.Action, requiredPerm.Limiter)
			if err != nil {
				slog.Error("error creating user permission", slog.String("error", err.Error()))
				continue
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "user app role added"})
}

func (h *HttpEndpoints) deleteManagementUserAppRole(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	userID := c.Param("userID")
	appRoleID := c.Param("appRoleID")

	slog.Info("deleting user app role", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedUserID", userID), slog.String("appRoleID", appRoleID))

	err := h.muDBConn.DeleteAppRole(token.InstanceID, appRoleID)
	if err != nil {
		slog.Error("error deleting user app role", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting user app role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user app role deleted"})
}

func (h *HttpEndpoints) getAllAppRoleTemplates(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("getting all app role templates", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	appRoleTemplates, err := h.muDBConn.GetAllAppRoleTemplates(token.InstanceID)
	if err != nil {
		slog.Error("error retrieving app role templates", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error retrieving app role templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appRoleTemplates": appRoleTemplates})
}

type AppRoleTemplateProps struct {
	AppName     string               `json:"appName"`
	Role        string               `json:"role"`
	Permissions []mUserDB.Permission `json:"permissions"`
}

func (h *HttpEndpoints) createAppRoleTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("creating app role template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	var req AppRoleTemplateProps
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.muDBConn.AddAppRoleTemplate(token.InstanceID, req.AppName, req.Role, req.Permissions)
	if err != nil {
		slog.Error("error creating app role template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating app role template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "app role template created"})
}

func (h *HttpEndpoints) getAppRoleTemplateByID(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	appRoleTemplateID := strings.TrimSpace(c.Param("appRoleTemplateID"))

	if appRoleTemplateID == "" {
		slog.Error("app role template ID is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "app role template ID is required"})
		return
	}

	slog.Info("getting app role template by ID", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("appRoleTemplateID", appRoleTemplateID))

	appRoleTemplate, err := h.muDBConn.GetAppRoleTemplateByID(token.InstanceID, appRoleTemplateID)
	if err != nil {
		slog.Error("error retrieving app role template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error retrieving app role template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appRoleTemplate": appRoleTemplate})
}

func (h *HttpEndpoints) updateAppRoleTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	appRoleTemplateID := strings.TrimSpace(c.Param("appRoleTemplateID"))

	if appRoleTemplateID == "" {
		slog.Error("app role template ID is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "app role template ID is required"})
		return
	}

	var req AppRoleTemplateProps
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("updating app role template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("appRoleTemplateID", appRoleTemplateID))

	err := h.muDBConn.UpdateAppRoleTemplate(token.InstanceID, appRoleTemplateID, req.AppName, req.Role, req.Permissions)
	if err != nil {
		slog.Error("error updating app role template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating app role template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "app role template updated"})
}

func (h *HttpEndpoints) deleteAppRoleTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	appRoleTemplateID := strings.TrimSpace(c.Param("appRoleTemplateID"))

	if appRoleTemplateID == "" {
		slog.Error("app role template ID is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "app role template ID is required"})
		return
	}

	slog.Info("deleting app role template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("appRoleTemplateID", appRoleTemplateID))

	err := h.muDBConn.DeleteAppRoleTemplate(token.InstanceID, appRoleTemplateID)
	if err != nil {
		slog.Error("error deleting app role template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting app role template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "app role template deleted"})
}

func (h *HttpEndpoints) deleteAppRoleTemplateForApp(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	appName := strings.TrimSpace(c.Param("appName"))

	if appName == "" {
		slog.Error("app name is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "app name is required"})
		return
	}

	slog.Info("deleting all app role templates (and its instance roles) for app", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("appName", appName))

	if err := h.muDBConn.RemoveAllAppRolesForApp(token.InstanceID, appName); err != nil {
		slog.Error("error deleting app roles for app", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting app roles for app"})
		return
	}

	if err := h.muDBConn.RemoveAllAppRoleTemplatesForApp(token.InstanceID, appName); err != nil {
		slog.Error("error deleting app role template for app", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting app role template for app"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "app role template deleted"})
}

func (h *HttpEndpoints) getAllAppRoles(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("getting all app roles", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	appRoles, err := h.muDBConn.GetAllAppRoles(token.InstanceID)
	if err != nil {
		slog.Error("error retrieving app roles", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error retrieving app roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appRoles": appRoles})
}

func (h *HttpEndpoints) deleteAllAppRolesForApp(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	appName := strings.TrimSpace(c.Param("appName"))

	if appName == "" {
		slog.Error("app name is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "app name is required"})
		return
	}

	slog.Info("deleting all app roles for app", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("appName", appName))

	err := h.muDBConn.RemoveAllAppRolesForApp(token.InstanceID, appName)
	if err != nil {
		slog.Error("error deleting app roles for app", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting app roles for app"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "app roles deleted"})
}

func (h *HttpEndpoints) requestParticipantUserDeletion(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !umUtils.CheckEmailFormat(req.Email) {
		slog.Error("invalid email format", slog.String("email", req.Email))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	slog.Info("requesting participant user deletion", slog.String("instanceID", token.InstanceID), slog.String("by", token.Subject), slog.String("email", req.Email))

	user, err := h.participantUserDB.GetUserByAccountID(token.InstanceID, req.Email)
	if err != nil {
		slog.Error("user not found", slog.String("instanceID", token.InstanceID), slog.String("email", req.Email), slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "user could not be deleted"})
		return
	}

	for _, profile := range user.Profiles {
		studyService.OnProfileDeleted(token.InstanceID, profile.ID.Hex(), nil)
	}

	// delete all temp tokens
	err = h.globalInfosDBConn.DeleteAllTempTokenForUser(token.InstanceID, user.ID.Hex(), "")
	if err != nil {
		slog.Error("failed to delete temp tokens", slog.String("error", err.Error()))
	}

	err = h.participantUserDB.DeleteUser(token.InstanceID, user.ID.Hex())
	if err != nil {
		slog.Error("cannot delete user", slog.String("instanceId", token.InstanceID), slog.String("userId", token.Subject), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})

}

func (h *HttpEndpoints) getAllServiceAccounts(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("getting all service accounts", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	serviceUser, err := h.muDBConn.GetServiceUsers(token.InstanceID)
	if err != nil {
		slog.Error("error retrieving service accounts", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting service accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"serviceAccounts": serviceUser})
}

type ServiceUserProps struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

func (h *HttpEndpoints) createServiceAccount(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	var req ServiceUserProps
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("creating a new service account", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("label", req.Label))

	nu, err := h.muDBConn.CreateServiceUser(token.InstanceID, req.Label, req.Description)
	if err != nil {
		slog.Error("error creating service account", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating service account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"serviceAccount": nu})
}

func (h *HttpEndpoints) getServiceAccount(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	slog.Info("getting service account", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("requestedServiceAccountID", serviceAccountID))

	serviceUser, err := h.muDBConn.GetServiceUserByID(token.InstanceID, serviceAccountID)
	if err != nil {
		slog.Error("error retrieving service account", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting service account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"serviceAccount": serviceUser})
}

func (h *HttpEndpoints) updateServiceAccount(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")
	var req ServiceUserProps
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("updating service account", slog.String("serviceAccountID", serviceAccountID))

	if err := h.muDBConn.UpdateServiceUser(token.InstanceID, serviceAccountID, req.Label, req.Description); err != nil {
		slog.Error("failed to update service account", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HttpEndpoints) getServiceAccountAPIKeys(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	slog.Info("getting API keys for service account", slog.String("serviceAccountID", serviceAccountID), slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	apiKeys, err := h.muDBConn.GetServiceUserAPIKeys(token.InstanceID, serviceAccountID)
	if err != nil {
		slog.Error("failed to get api keys for service account", slog.String("error", err.Error()), slog.String("instanceID", token.InstanceID), slog.String("serviceAccountID", serviceAccountID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get API keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apiKeys": apiKeys,
	})
}

type ServiceAccountAPIKeyRequest struct {
	ExpiresAt int64 `json:"expiresAt,omitempty"`
}

func (h *HttpEndpoints) createServiceAccountAPIKey(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	var req ServiceAccountAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("creating service account API key", slog.String("serviceAccountID", serviceAccountID), slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	_, err := h.muDBConn.GetServiceUserByID(token.InstanceID, serviceAccountID)
	if err != nil {
		slog.Error("service account not found", slog.String("serviceAccountID", serviceAccountID), slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "service account not found"})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt > 0 {
		eat := time.Unix(req.ExpiresAt, 0)
		expiresAt = &eat
	}

	newApiKey, err := utils.GenerateUniqueTokenString()
	if err != nil {
		slog.Error("failed to generate unique token string", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.muDBConn.CreateServiceUserAPIKey(token.InstanceID, serviceAccountID, newApiKey, expiresAt); err != nil {
		slog.Error("failed to create service account API key", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HttpEndpoints) deleteServiceAccountAPIKey(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")
	apiKeyID := c.Param("apiKeyID")

	slog.Info("deleting service account API key", slog.String("serviceAccountID", serviceAccountID), slog.String("apiKeyID", apiKeyID), slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	if err := h.muDBConn.DeleteServiceUserAPIKey(token.InstanceID, apiKeyID); err != nil {
		slog.Error("failed to delete service account API key", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HttpEndpoints) deleteServiceAccount(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	slog.Info("deleting service account", slog.String("serviceAccountID", serviceAccountID), slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	if err := h.muDBConn.DeleteServiceUser(token.InstanceID, serviceAccountID); err != nil {
		slog.Error("failed to delete service account", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HttpEndpoints) getServiceAccountPermissions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	slog.Info("getting user permissions", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("serviceAccountID", serviceAccountID))

	permissions, err := h.muDBConn.GetPermissionBySubject(token.InstanceID, serviceAccountID, pc.SUBJECT_TYPE_SERVICE_ACCOUNT)
	if err != nil {
		slog.Error("error retrieving sercice account permissions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting service account permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

func (h *HttpEndpoints) createServiceAccountPermission(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	var newPerm mUserDB.Permission
	if err := c.ShouldBindJSON(&newPerm); err != nil {
		slog.Error("error binding permission", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing payload"})
		return
	}

	slog.Info("creating service account permission", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("serviceAccountID", serviceAccountID))

	_, err := h.muDBConn.GetServiceUserByID(token.InstanceID, serviceAccountID)
	if err != nil {
		slog.Error("service account not found", slog.String("serviceAccountID", serviceAccountID), slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))
		c.JSON(http.StatusBadRequest, gin.H{"error": "service account not found"})
		return
	}

	newPerm.SubjectType = pc.SUBJECT_TYPE_SERVICE_ACCOUNT
	newPerm.SubjectID = serviceAccountID

	permission, err := h.muDBConn.CreatePermission(
		token.InstanceID,
		newPerm.SubjectID,
		newPerm.SubjectType,
		newPerm.ResourceType,
		newPerm.ResourceKey,
		newPerm.Action,
		newPerm.Limiter,
	)
	if err != nil {
		slog.Error("error creating service account permission", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating service account permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permission": permission})
}

func (h *HttpEndpoints) deleteServiceAccountPermission(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	permissionID := c.Param("permissionID")

	slog.Info("deleting service account permission", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("serviceAccountID", serviceAccountID), slog.String("permissionID", permissionID))

	err := h.muDBConn.DeletePermission(token.InstanceID, permissionID)
	if err != nil {
		slog.Error("error deleting service account permission", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting service account permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission deleted"})
}

func (h *HttpEndpoints) updateServiceAccountPermissionLimiter(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	serviceAccountID := c.Param("serviceAccountID")

	permissionID := c.Param("permissionID")

	var newLimiter mUserDB.Permission
	if err := c.ShouldBindJSON(&newLimiter); err != nil {
		slog.Error("error binding permission", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing payload"})
		return
	}

	slog.Info("updating service account permission limiter", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("serviceAccountID", serviceAccountID), slog.String("permissionID", permissionID))

	err := h.muDBConn.UpdatePermissionLimiter(token.InstanceID, permissionID, newLimiter.Limiter)
	if err != nil {
		slog.Error("error updating service account permission limiter", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating service account permission limiter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission limiter updated"})
}
