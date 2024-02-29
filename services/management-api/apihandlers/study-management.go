package apihandlers

import (
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	pc "github.com/case-framework/case-backend/pkg/permission-checker"
	"github.com/gin-gonic/gin"
)

func (h *HttpEndpoints) AddStudyManagementAPI(rg *gin.RouterGroup) {
	studiesGroup := rg.Group("/studies")

	studiesGroup.Use(mw.GetAndValidateManagementUserJWT(h.tokenSignKey))
	studiesGroup.Use(mw.IsInstanceIDInJWTAllowed(h.allowedInstanceIDs))
	{
		studiesGroup.GET("/", h.getAllStudies)
		studiesGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType: pc.RESOURCE_TYPE_STUDY,
				ResourceKeys: []string{pc.RESOURCE_KEY_STUDY_ALL},
				Action:       pc.ACTION_CREATE_STUDY,
			},
			nil,
			h.createStudy,
		))
	}

	// Study Group
	studyGroup := studiesGroup.Group("/:studyKey")
	{
		h.addGeneralStudyEndpoints(studyGroup)
		h.addStudyConfigEndpoints(studyGroup)
		h.addStudyRuleEndpoints(studyGroup)
		h.addSurveyEndpoints(studyGroup)
		// h.addStudyActionEndpoints(studyGroup) --> TODO: define async task based actions

		// TODO: study permissions
	}
}

func getStudyKeyFromParams(c *gin.Context) []string {
	return []string{c.Param("studyKey")}
}

func getSurveyKeyLimiterFromContext(c *gin.Context) map[string]string {
	return map[string]string{"surveyKey": c.Param("surveyKey")}
}

func (h *HttpEndpoints) addGeneralStudyEndpoints(rg *gin.RouterGroup) {
	rg.GET("/", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_READ_STUDY_CONFIG,
		},
		nil,
		h.getStudyProps,
	))

	rg.PUT("/", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_PROPS,
		},
		nil,
		h.updateStudyProps,
	))

	// change status
	rg.PUT("/status", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_STATUS,
		},
		nil,
		h.updateStudyStatus,
	))

	rg.DELETE("/", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_DELETE_STUDY,
		},
		nil,
		h.deleteStudy,
	))
}

func (h *HttpEndpoints) addSurveyEndpoints(rg *gin.RouterGroup) {
	surveysGroup := rg.Group("/surveys")
	{
		surveysGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_READ_STUDY_CONFIG,
			},
			nil,
			h.getSurveyInfoList,
		))

		surveysGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_CREATE_SURVEY,
			},
			nil,
			h.createSurvey,
		))
	}

	surveyGroup := surveysGroup.Group("/:surveyKey")
	{
		surveyGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_READ_STUDY_CONFIG,
			},
			nil,
			h.getLatestSurvey,
		))

		surveyGroup.PUT("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_UPDATE_SURVEY,
			},
			getSurveyKeyLimiterFromContext,
			h.updateSurvey,
		))

		surveyGroup.DELETE("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_UNPUBLISH_SURVEY,
			},
			getSurveyKeyLimiterFromContext,
			h.unpublishSurvey,
		))

		surveyGroup.GET("/versions", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_READ_STUDY_CONFIG,
			},
			nil,
			h.getSurveyVersions,
		))

		surveyGroup.GET("/versions/:versionID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_READ_STUDY_CONFIG,
			},
			nil,
			h.getSurveyVersion,
		))

		surveyGroup.DELETE("/versions/:versionID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_DELETE_SURVEY_VERSION,
			},
			getSurveyKeyLimiterFromContext,
			h.deleteSurveyVersion,
		))

	}
}

func (h *HttpEndpoints) addStudyConfigEndpoints(rg *gin.RouterGroup) {

	permissionsGroup := rg.Group("/permissions")
	{
		permissionsGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_READ_STUDY_CONFIG,
			},
			nil,
			h.getStudyPermissions,
		))

		permissionsGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_MANAGE_STUDY_PERMISSIONS,
			},
			nil,
			h.addStudyPermission,
		))

		permissionsGroup.PUT("/:permissionID", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_MANAGE_STUDY_PERMISSIONS,
			},
			nil,
			h.updateStudyPermissions,
		))

		permissionsGroup.DELETE("/:permissionID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_MANAGE_STUDY_PERMISSIONS,
			},
			nil,
			h.deleteStudyPermission,
		))
	}

	notificationSubGroup := rg.Group("/notification-subscriptions")
	{
		notificationSubGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_READ_STUDY_CONFIG,
			},
			nil,
			h.getNotificationSubscriptions,
		))

		notificationSubGroup.PUT("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_UPDATE_NOTIFICATION_SUBSCRIPTIONS,
			},
			nil,
			h.updateNotificationSubscriptions,
		))
	}
}

func (h *HttpEndpoints) addStudyRuleEndpoints(rg *gin.RouterGroup) {
	rulesGroup := rg.Group("/rules")

	rulesGroup.GET("/", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_READ_STUDY_CONFIG,
		},
		nil,
		h.getStudyRules,
	))

	rulesGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_RULES,
		},
		nil,
		h.updateStudyRules,
	))

	// get rule history
	rulesGroup.GET("/versions", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_READ_STUDY_CONFIG,
		},
		nil,
		h.getStudyRuleVersions,
	))

	// get specific rule version
	rulesGroup.GET("/versions/:versionID", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_READ_STUDY_CONFIG,
		},
		nil,
		h.getStudyRuleVersion,
	))

	// delete rule version
	rulesGroup.DELETE("/versions/:versionID", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_RULES,
		},
		nil,
		h.deleteStudyRuleVersion,
	))

}

func (h *HttpEndpoints) getAllStudies(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) createStudy(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyProps(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) updateStudyProps(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) updateStudyStatus(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteStudy(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getSurveyInfoList(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) createSurvey(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getLatestSurvey(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) updateSurvey(c *gin.Context) {
	//	TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) unpublishSurvey(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getSurveyVersions(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getSurveyVersion(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteSurveyVersion(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyPermissions(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) addStudyPermission(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) updateStudyPermissions(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteStudyPermission(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getNotificationSubscriptions(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) updateNotificationSubscriptions(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyRules(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) updateStudyRules(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyRuleVersions(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyRuleVersion(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteStudyRuleVersion(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
