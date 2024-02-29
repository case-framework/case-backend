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
		h.addStudyActionEndpoints(studyGroup)
		h.addStudyDataExporterEndpoints(studyGroup)
		h.addStudyDataExplorerEndpoints(studyGroup)
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

func (h *HttpEndpoints) addStudyActionEndpoints(rg *gin.RouterGroup) {
	actionsGroup := rg.Group("/run-actions")

	// run actions on current participant states
	participantGroup := actionsGroup.Group("/participants")
	{
		participantGroup.POST("/:participantID", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.runActionOnParticipant,
		))

		participantGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.runActionOnParticipants,
		))

		participantGroup.GET("/task/:taskID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.getStudyActionTaskStatus,
		))

		participantGroup.GET("/task/:taskID/result", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.getStudyActionTaskResult,
		))
	}

	// run action on previous responses of a participant
	previousResponsesGroup := actionsGroup.Group("/previous-responses")
	{
		previousResponsesGroup.POST("/:participantID", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.runActionOnPreviousResponsesForParticipant,
		))

		previousResponsesGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.runActionOnPreviousResponsesForParticipants,
		))

		previousResponsesGroup.GET("/task/:taskID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.getStudyActionTaskStatus,
		))

		previousResponsesGroup.GET("/task/:taskID/result", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_RUN_STUDY_ACTION,
			},
			nil,
			h.getStudyActionTaskResult,
		))
	}
}

func (h *HttpEndpoints) addStudyDataExporterEndpoints(rg *gin.RouterGroup) {
	exporterGroup := rg.Group("/data-exporter")

	responsesGroup := exporterGroup.Group("/responses")
	{
		// count responses
		responsesGroup.GET("/count", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_RESPONSES,
			},
			nil,
			h.getResponsesCount,
		))

		// start export generation for responses
		responsesGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_RESPONSES,
			},
			nil,
			h.generateResponsesExport,
		))

		// get export status
		responsesGroup.GET("/task/:taskID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_RESPONSES,
			},
			nil,
			h.getExportTaskStatus,
		))

		// get export result
		responsesGroup.GET("/task/:taskID/result", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_RESPONSES,
			},
			nil,
			h.getExportTaskResult,
		))
	}

	participantsGroup := exporterGroup.Group("/participants")
	{
		// count participants of the query
		participantsGroup.GET("/count", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_PARTICIPANT_STATES,
			},
			nil,
			h.getParticipantsCount,
		))

		// start export generation for participants
		participantsGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_PARTICIPANT_STATES,
			},
			nil,
			h.generateParticipantsExport,
		))

		// get export status
		participantsGroup.GET("/task/:taskID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_PARTICIPANT_STATES,
			},
			nil,
			h.getExportTaskStatus,
		))

		// get export result
		participantsGroup.GET("/task/:taskID/result", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_PARTICIPANT_STATES,
			},
			nil,
			h.getExportTaskResult,
		))
	}

	reportsGroup := exporterGroup.Group("/reports")
	{
		// count reports
		reportsGroup.GET("/count", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_REPORTS,
			},
			nil,
			h.getReportsCount,
		))

		// start export generation for reports
		reportsGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_REPORTS,
			},
			nil,
			h.generateReportsExport,
		))

		// get export status
		reportsGroup.GET("/task/:taskID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_REPORTS,
			},
			nil,
			h.getExportTaskStatus,
		))

		// get export result
		reportsGroup.GET("/task/:taskID/result", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_REPORTS,
			},
			nil,
			h.getExportTaskResult,
		))
	}

	confidentialResponsesGroup := exporterGroup.Group("/confidential-responses")
	{
		// count confidential responses
		confidentialResponsesGroup.GET("/count", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_CONFIDENTIAL_RESPONSES,
			},
			nil,
			h.getConfidentialResponsesCount,
		))

		// start export generation for confidential responses
		confidentialResponsesGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_CONFIDENTIAL_RESPONSES,
			},
			nil,
			h.generateConfidentialResponsesExport,
		))

		// get export status
		confidentialResponsesGroup.GET("/task/:taskID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_CONFIDENTIAL_RESPONSES,
			},
			nil,
			h.getExportTaskStatus,
		))

		// get export result
		confidentialResponsesGroup.GET("/task/:taskID/result", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_CONFIDENTIAL_RESPONSES,
			},
			nil,
			h.getExportTaskResult,
		))
	}
}

func (h *HttpEndpoints) addStudyDataExplorerEndpoints(rg *gin.RouterGroup) {
	dataExplGroup := rg.Group("/data-explorer")

	responsesGroup := dataExplGroup.Group("/responses")
	{
		// get responses with pagination
		responsesGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_RESPONSES,
			},
			nil,
			h.getStudyResponses,
		))

		// delete responses
		responsesGroup.DELETE("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_DELETE_RESPONSES,
			},
			nil,
			h.deleteStudyResponses,
		))
	}

	participantsGroup := dataExplGroup.Group("/participants")
	{
		// get participants with pagination
		participantsGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_PARTICIPANT_STATES,
			},
			nil,
			h.getStudyParticipants,
		))

		// get single participant
		participantsGroup.GET("/:participantID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_PARTICIPANT_STATES,
			},
			nil,
			h.getStudyParticipant,
		))
	}

	reportsGroup := dataExplGroup.Group("/reports")
	{
		// get reports with pagination
		reportsGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_REPORTS,
			},
			nil,
			h.getStudyReports,
		))

		// get single report by ID
		reportsGroup.GET("/:reportID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_REPORTS,
			},
			nil,
			h.getStudyReport,
		))
	}

	filesGroup := dataExplGroup.Group("/files")
	{
		// get files with pagination
		filesGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_FILES,
			},
			nil,
			h.getStudyFiles,
		))

		// get single file by ID
		filesGroup.GET("/:fileID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_FILES,
			},
			nil,
			h.getStudyFile,
		))

		// delete file by ID
		filesGroup.DELETE("/:fileID", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_DELETE_FILES,
			},
			nil,
			h.deleteStudyFile,
		))
	}
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

func (h *HttpEndpoints) runActionOnParticipant(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) runActionOnParticipants(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyActionTaskStatus(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyActionTaskResult(c *gin.Context) {
	// TODO: implement
	// TODO: cleanup after successfully retrieving results
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) runActionOnPreviousResponsesForParticipant(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) runActionOnPreviousResponsesForParticipants(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getResponsesCount(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) generateResponsesExport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getParticipantsCount(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) generateParticipantsExport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getReportsCount(c *gin.Context) {
	//	TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) generateReportsExport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getConfidentialResponsesCount(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) generateConfidentialResponsesExport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getExportTaskStatus(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getExportTaskResult(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyResponses(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteStudyResponses(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyParticipants(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyParticipant(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyReports(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyReport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyFiles(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyFile(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteStudyFile(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
