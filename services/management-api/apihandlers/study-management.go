package apihandlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/case-framework/case-backend/pkg/apihelpers"
	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	managementuser "github.com/case-framework/case-backend/pkg/db/management-user"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	pc "github.com/case-framework/case-backend/pkg/permission-checker"
	"github.com/case-framework/case-backend/pkg/utils"
	"github.com/gin-gonic/gin"

	studydefinition "github.com/case-framework/case-backend/pkg/exporter/survey-definition"
	surveydefinition "github.com/case-framework/case-backend/pkg/exporter/survey-definition"
	surveyresponses "github.com/case-framework/case-backend/pkg/exporter/survey-responses"
	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
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

func getSurveyKeyLimiterFromQuery(c *gin.Context) map[string]string {
	sk := c.DefaultQuery("surveyKey", "")
	if sk == "" {
		return nil
	}
	return map[string]string{"surveyKey": sk}
}

func getReportKeyLimiterFromQuery(c *gin.Context) map[string]string {
	rk := c.DefaultQuery("reportKey", "")
	if rk == "" {
		return nil
	}
	return map[string]string{"reportKey": rk}
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

	rg.PUT("/is-default", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_PROPS,
		},
		nil,
		h.updateStudyIsDefault,
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

	// update study display props (name, description, tags)
	rg.PUT("/display-props", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_PROPS,
		},
		nil,
		h.updateStudyDisplayProps,
	))

	rg.PUT("/file-upload-config", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_PROPS,
		},
		nil,
		h.updateStudyFileUploadRule,
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

		surveyGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_UPDATE_SURVEY,
			},
			getSurveyKeyLimiterFromContext,
			h.updateSurvey,
		))

		surveyGroup.POST("/unpublish", h.useAuthorisedHandler(
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
		h.getCurrentStudyRules,
	))

	rulesGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_UPDATE_STUDY_RULES,
		},
		nil,
		h.publishNewStudyRulesVersion,
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
	rulesGroup.GET("/versions/:id", h.useAuthorisedHandler(
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
	rulesGroup.DELETE("/versions/:id", h.useAuthorisedHandler(
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

	surveyInfoGroup := exporterGroup.Group("/survey-info")
	{
		// get survey info
		surveyInfoGroup.GET("/", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_READ_STUDY_CONFIG,
			},
			nil,
			h.getSurveyInfo,
		))
	}

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
			getSurveyKeyLimiterFromQuery,
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
			getSurveyKeyLimiterFromQuery,
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
			getReportKeyLimiterFromQuery,
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
			getReportKeyLimiterFromQuery,
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

		// delete confidential response
		confidentialResponsesGroup.GET("/:id", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_CONFIDENTIAL_RESPONSES,
			},
			nil,
			h.deleteConfidentialResponse,
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
			getSurveyKeyLimiterFromQuery,
			h.getStudyResponses,
		))

		responsesGroup.GET("/:responseId", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_GET_RESPONSES,
			},
			getSurveyKeyLimiterFromQuery,
			h.getStudyResponseById,
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

		responsesGroup.DELETE("/:responseId", h.useAuthorisedHandler(
			RequiredPermission{
				ResourceType:        pc.RESOURCE_TYPE_STUDY,
				ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
				ExtractResourceKeys: getStudyKeyFromParams,
				Action:              pc.ACTION_DELETE_RESPONSES,
			},
			nil,
			h.deleteStudyResponse,
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
			getReportKeyLimiterFromQuery,
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
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("getting all studies", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	studies, err := h.studyDBConn.GetStudies(token.InstanceID, "", false)
	if err != nil {
		slog.Error("failed to get all studies", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get studies"})
		return
	}

	for i := range studies {
		studies[i].SecretKey = ""
		studies[i].Rules = nil
		studies[i].NotificationSubscriptions = nil
	}

	c.JSON(http.StatusOK, gin.H{"studies": studies})
}

type NewStudyReq struct {
	StudyKey             string `json:"studyKey"`
	SecretKey            string `json:"secretKey"`
	IsSystemDefaultStudy bool   `json:"isSystemDefaultStudy"`
}

func (h *HttpEndpoints) createStudy(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	var req NewStudyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("creating new study", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", req.StudyKey))

	// check if study key is URL safe
	if !utils.IsURLSafe(req.StudyKey) {
		slog.Error("study key is not URL safe", slog.String("studyKey", req.StudyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "study key is not URL safe"})
		return
	}

	study := studyTypes.Study{
		Key:       req.StudyKey,
		SecretKey: req.SecretKey,
		Status:    studyTypes.STUDY_STATUS_INACTIVE,
		Props: studyTypes.StudyProps{
			SystemDefaultStudy: req.IsSystemDefaultStudy,
		},
		Configs: studyTypes.StudyConfigs{
			IdMappingMethod: studyTypes.DEFAULT_ID_MAPPING_METHOD,
			ParticipantFileUploadRule: &studyTypes.Expression{
				Name: "gt",
				Data: []studyTypes.ExpressionArg{
					{Num: 0, DType: "num"},
					{Num: 1, DType: "num"},
				},
			}, // default rule: file upload is not allowed
		},
	}

	err := h.studyDBConn.CreateStudy(token.InstanceID, study)
	if err != nil {
		slog.Error("failed to create study", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create study"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"study": study})
}

func (h *HttpEndpoints) getStudyProps(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	slog.Info("getting study props", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	study, err := h.studyDBConn.GetStudy(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("failed to get study", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"study": study})
}

type StudyIsDefaultUpdateReq struct {
	IsDefault bool `json:"isDefault"`
}

func (h *HttpEndpoints) updateStudyIsDefault(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var req StudyIsDefaultUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("updating study is default", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.Bool("isDefault", req.IsDefault))

	err := h.studyDBConn.UpdateStudyIsDefault(token.InstanceID, studyKey, req.IsDefault)
	if err != nil {
		slog.Error("failed to update study is default", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update study is default"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study is default updated"})
}

type StudyStatusUpdateReq struct {
	Status string `json:"status"`
}

func (h *HttpEndpoints) updateStudyStatus(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var req StudyStatusUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("updating study status", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("status", req.Status))

	err := h.studyDBConn.UpdateStudyStatus(token.InstanceID, studyKey, req.Status)
	if err != nil {
		slog.Error("failed to update study status", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update study status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "study status updated"})
}

type StudyDisplayPropsUpdateReq struct {
	Name        []studyTypes.LocalisedObject `bson:"name" json:"name"`
	Description []studyTypes.LocalisedObject `bson:"description" json:"description"`
	Tags        []studyTypes.Tag             `bson:"tags" json:"tags"`
}

func (h *HttpEndpoints) updateStudyDisplayProps(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var req StudyDisplayPropsUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("updating study display props", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	err := h.studyDBConn.UpdateStudyDisplayProps(token.InstanceID, studyKey, req.Name, req.Description, req.Tags)
	if err != nil {
		slog.Error("failed to update study display props", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update study display props"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "study display props updated"})
}

type FileUploadRuleUpdateReq struct {
	SimplifiedAllow bool                   `json:"simplifiedAllowedUpload"`
	Expression      *studyTypes.Expression `json:"expression,omitempty"`
}

func (h *HttpEndpoints) updateStudyFileUploadRule(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var req FileUploadRuleUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	newRule := req.Expression
	if newRule == nil {
		if req.SimplifiedAllow {
			newRule = &studyTypes.Expression{
				Name: "gt",
				Data: []studyTypes.ExpressionArg{
					{Num: 1, DType: "num"},
					{Num: 0, DType: "num"},
				},
			}
		} else {
			newRule = &studyTypes.Expression{
				Name: "gt",
				Data: []studyTypes.ExpressionArg{
					{Num: 0, DType: "num"},
					{Num: 1, DType: "num"},
				},
			}
		}
	}

	slog.Info("updating study file upload rule", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	err := h.studyDBConn.UpdateStudyFileUploadRule(token.InstanceID, studyKey, newRule)
	if err != nil {
		slog.Error("failed to update study file upload rule", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update study file upload rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "study file upload rule updated"})
}

func (h *HttpEndpoints) deleteStudy(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	slog.Info("deleting study", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	err := h.studyDBConn.DeleteStudy(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("failed to delete study", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete study"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study deleted"})
}

type SurveyInfo struct {
	Key string `json:"key"`
}

func (h *HttpEndpoints) getSurveyInfoList(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	slog.Info("getting survey info list", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	surveyKeys, err := h.studyDBConn.GetSurveyKeysForStudy(token.InstanceID, studyKey, true)
	if err != nil {
		slog.Error("failed to get survey info list", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey info list"})
		return
	}

	surveyInfos := make([]SurveyInfo, len(surveyKeys))
	for i, key := range surveyKeys {
		surveyInfos[i] = SurveyInfo{Key: key}
	}

	c.JSON(http.StatusOK, gin.H{"surveys": surveyInfos})
}

func (h *HttpEndpoints) createSurvey(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var survey studyTypes.Survey
	if err := c.ShouldBindJSON(&survey); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	survey.SurveyKey = survey.SurveyDefinition.Key

	slog.Info("creating survey", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", survey.SurveyDefinition.Key))

	surveyKeys, err := h.studyDBConn.GetSurveyKeysForStudy(token.InstanceID, studyKey, true)
	if err != nil {
		slog.Error("failed to get survey info list", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey info list"})
		return
	}

	for _, key := range surveyKeys {
		if key == survey.SurveyKey {
			slog.Error("survey key already exists", slog.String("key", survey.SurveyKey))
			c.JSON(http.StatusBadRequest, gin.H{"error": "survey key already exists"})
			return
		}
	}

	if survey.VersionID == "" {
		surveyHistory, err := h.studyDBConn.GetSurveyVersions(token.InstanceID, studyKey, survey.SurveyKey)
		if err != nil {
			slog.Error("failed to get survey versions", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey versions"})
			return
		}
		survey.VersionID = utils.GenerateSurveyVersionID(surveyHistory)
	}

	survey.Published = time.Now().Unix()

	err = h.studyDBConn.SaveSurveyVersion(token.InstanceID, studyKey, &survey)
	if err != nil {
		slog.Error("failed to create survey", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create survey"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"survey": survey})
}

func (h *HttpEndpoints) getLatestSurvey(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	surveyKey := c.Param("surveyKey")

	slog.Info("getting latest survey", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))

	survey, err := h.studyDBConn.GetCurrentSurveyVersion(token.InstanceID, studyKey, surveyKey)
	if err != nil {
		slog.Error("failed to get latest survey", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get latest survey"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"survey": survey})
}

func (h *HttpEndpoints) updateSurvey(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	surveyKey := c.Param("surveyKey")

	var survey studyTypes.Survey
	if err := c.ShouldBindJSON(&survey); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	survey.SurveyKey = survey.SurveyDefinition.Key

	if survey.SurveyKey != surveyKey {
		slog.Error("survey key in request does not match", slog.String("key", survey.SurveyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "survey key in request does not match"})
		return
	}

	if survey.VersionID == "" {
		surveyHistory, err := h.studyDBConn.GetSurveyVersions(token.InstanceID, studyKey, survey.SurveyKey)
		if err != nil {
			slog.Error("failed to get survey versions", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey versions"})
			return
		}
		survey.VersionID = utils.GenerateSurveyVersionID(surveyHistory)
	}

	survey.Published = time.Now().Unix()

	slog.Info("updating survey", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))

	err := h.studyDBConn.SaveSurveyVersion(token.InstanceID, studyKey, &survey)
	if err != nil {
		slog.Error("failed to update survey", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update survey"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"survey": survey})
}

func (h *HttpEndpoints) unpublishSurvey(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	surveyKey := c.Param("surveyKey")

	slog.Info("unpublishing survey", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))

	err := h.studyDBConn.UnpublishSurvey(token.InstanceID, studyKey, surveyKey)
	if err != nil {
		slog.Error("failed to unpublish survey", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unpublish survey"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "survey unpublished"})
}

func (h *HttpEndpoints) getSurveyVersions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	surveyKey := c.Param("surveyKey")

	slog.Info("getting survey versions", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))

	versions, err := h.studyDBConn.GetSurveyVersions(token.InstanceID, studyKey, surveyKey)
	if err != nil {
		slog.Error("failed to get survey versions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey versions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

func (h *HttpEndpoints) getSurveyVersion(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	surveyKey := c.Param("surveyKey")
	versionID := c.Param("versionID")

	slog.Info("getting survey version", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey), slog.String("versionID", versionID))

	version, err := h.studyDBConn.GetSurveyVersion(token.InstanceID, studyKey, surveyKey, versionID)

	if err != nil {
		slog.Error("failed to get survey version", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"survey": version})
}

func (h *HttpEndpoints) deleteSurveyVersion(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	surveyKey := c.Param("surveyKey")
	versionID := c.Param("versionID")

	slog.Info("deleting survey version", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey), slog.String("versionID", versionID))

	err := h.studyDBConn.DeleteSurveyVersion(token.InstanceID, studyKey, surveyKey, versionID)
	if err != nil {
		slog.Error("failed to delete survey version", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete survey version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "survey version deleted"})
}

type StudyUserPermissionInfo struct {
	User        *managementuser.ManagementUser `json:"user"`
	Permissions []managementuser.Permission    `json:"permissions"`
}

func (h *HttpEndpoints) getStudyPermissions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	permissions, err := h.muDBConn.GetPermissionByResource(token.InstanceID, pc.RESOURCE_TYPE_STUDY, studyKey)
	if err != nil {
		slog.Error("failed to get study permissions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study permissions"})
		return
	}

	// check if user has "manage study permissions" permission
	// or is admin
	allowedToManagePermissions := false
	if token.IsAdmin {
		allowedToManagePermissions = true
	} else {
		for _, permission := range permissions {
			if permission.SubjectID == token.Subject &&
				permission.SubjectType == pc.SUBJECT_TYPE_MANAGEMENT_USER &&
				permission.Action == pc.ACTION_MANAGE_STUDY_PERMISSIONS {
				allowedToManagePermissions = true
				break
			}
		}
	}

	studyUserPermissionInfos := map[string]*StudyUserPermissionInfo{}

	for _, permission := range permissions {
		userID := permission.SubjectID

		if permission.SubjectType != pc.SUBJECT_TYPE_MANAGEMENT_USER {
			continue
		}

		var user *managementuser.ManagementUser

		// Check if user ID already exists in the map
		_, ok := studyUserPermissionInfos[userID]
		if !ok {
			// Get user info
			var err error
			user, err = h.muDBConn.GetUserByID(token.InstanceID, permission.SubjectID)
			if err != nil {
				slog.Error("failed to get user info", slog.String("error", err.Error()))
				continue
			}
			studyUserPermissionInfos[userID] = &StudyUserPermissionInfo{
				User: &managementuser.ManagementUser{
					ID:       user.ID,
					Username: user.Username,
					Email:    user.Email,
					ImageURL: user.ImageURL,
				},
				Permissions: []managementuser.Permission{},
			}
		}

		if allowedToManagePermissions {
			studyUserPermissionInfos[userID].Permissions = append(studyUserPermissionInfos[userID].Permissions, *permission)
		}
	}

	c.JSON(http.StatusOK, gin.H{"permissions": studyUserPermissionInfos})
}

func (h *HttpEndpoints) addStudyPermission(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var permission managementuser.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("adding study permission", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("subjectID", permission.SubjectID), slog.String("action", permission.Action))

	permission.SubjectType = pc.SUBJECT_TYPE_MANAGEMENT_USER
	permission.ResourceType = pc.RESOURCE_TYPE_STUDY
	permission.ResourceKey = studyKey

	_, err := h.muDBConn.CreatePermission(
		token.InstanceID,
		permission.SubjectID,
		permission.SubjectType,
		permission.ResourceType,
		permission.ResourceKey,
		permission.Action,
		permission.Limiter,
	)

	if err != nil {
		slog.Error("failed to add study permission", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add study permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study permission added"})
}

func (h *HttpEndpoints) deleteStudyPermission(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	permissionID := c.Param("permissionID")

	slog.Info("deleting study permission", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("permissionID", permissionID))

	permission, err := h.muDBConn.GetPermissionByID(token.InstanceID, permissionID)
	if err != nil {
		slog.Error("failed to get study permission", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study permission"})
		return
	}

	if permission.ResourceType != pc.RESOURCE_TYPE_STUDY || permission.ResourceKey != studyKey {
		slog.Warn("permission does not belong to the study", slog.String("permissionID", permissionID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission does not belong to the study"})
		return
	}

	err = h.muDBConn.DeletePermission(token.InstanceID, permissionID)
	if err != nil {
		slog.Error("failed to delete study permission", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete study permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study permission deleted"})
}

func (h *HttpEndpoints) getNotificationSubscriptions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	slog.Info("getting notification subscriptions", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	subscriptions, err := h.studyDBConn.GetNotificationSubscriptions(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("failed to get notification subscriptions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notification subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

type NotificationSubscriptionsUpdateReq struct {
	Subscriptions []studyTypes.NotificationSubscription `json:"subscriptions"`
}

func (h *HttpEndpoints) updateNotificationSubscriptions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var req NotificationSubscriptionsUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("updating notification subscriptions", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	err := h.studyDBConn.UpdateStudyNotificationSubscriptions(token.InstanceID, studyKey, req.Subscriptions)
	if err != nil {
		slog.Error("failed to update notification subscriptions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification subscriptions updated"})
}

func (h *HttpEndpoints) getCurrentStudyRules(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	slog.Info("getting current study rules", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	rules, err := h.studyDBConn.GetCurrentStudyRules(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("failed to get current study rules", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current study rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"studyRules": rules})
}

func (h *HttpEndpoints) publishNewStudyRulesVersion(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	var rules studyTypes.StudyRules
	if err := c.ShouldBindJSON(&rules); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	rules.StudyKey = studyKey
	rules.UploadedAt = time.Now().Unix()
	rules.UploadedBy = token.Subject

	err := rules.MarshalRules()
	if err != nil {
		slog.Error("failed to marshal study rules", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid study rules"})
		return
	}

	slog.Info("publishing new study rules version", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	err = h.studyDBConn.SaveStudyRules(token.InstanceID, studyKey, rules)
	if err != nil {
		slog.Error("failed to publish new study rules version", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish new study rules version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "new study rules version published"})
}

func (h *HttpEndpoints) getStudyRuleVersions(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	slog.Info("getting study rule versions", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	versions, err := h.studyDBConn.GetStudyRulesHistory(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("failed to get study rule versions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study rule versions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

func (h *HttpEndpoints) getStudyRuleVersion(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	versionID := c.Param("id")

	slog.Info("getting study rule version", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("versionID", versionID))

	version, err := h.studyDBConn.GetStudyRulesByID(token.InstanceID, studyKey, versionID)
	if err != nil {
		slog.Error("failed to get study rule version", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study rule version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"studyRules": version})
}

func (h *HttpEndpoints) deleteStudyRuleVersion(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	versionID := c.Param("id")

	slog.Info("deleting study rule version", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("versionID", versionID))

	err := h.studyDBConn.DeleteStudyRulesByID(token.InstanceID, studyKey, versionID)
	if err != nil {
		slog.Error("failed to delete study rule version", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete study rule version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study rule version deleted"})
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

func (h *HttpEndpoints) getSurveyInfo(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	surveyKey := c.DefaultQuery("surveyKey", "")
	if surveyKey == "" {
		slog.Error("surveyKey is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "surveyKey is required"})
		return
	}

	format := c.DefaultQuery("format", "json")
	if format != "json" && format != "csv" {
		slog.Error("invalid format", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey), slog.String("format", format))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format query parameter"})
		return
	}
	// includeItems, excludeItems
	language := c.DefaultQuery("language", "en")
	shortKeys := c.DefaultQuery("shortKeys", "false") == "true"

	slog.Info("getting survey info", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))

	sInfos, err := studydefinition.PrepareSurveyInfosFromDB(
		h.studyDBConn,
		token.InstanceID,
		studyKey,
		surveyKey,
		&surveydefinition.ExtractOptions{
			UseLabelLang: language,
		},
	)
	if err != nil {
		slog.Error("failed to get survey info", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey info"})
		return
	}

	siExp := studydefinition.NewSurveyInfoExporter(
		sInfos,
		surveyKey,
		shortKeys,
	)

	if format == "json" {
		c.Header("Content-Disposition", `attachment; filename=`+fmt.Sprintf("survey-infos_%s_%s.json", studyKey, surveyKey))
		c.JSON(http.StatusOK, gin.H{"versions": siExp.GetSurveyInfos(), "key": surveyKey})
		return
	}

	// CSV:
	c.Header("Content-Disposition", `attachment; filename=`+fmt.Sprintf("survey-infos_%s_%s.csv", studyKey, surveyKey))
	c.Header("Content-Type", "text/csv")
	err = siExp.GetSurveyInfoCSV(c.Writer)
	if err != nil {
		slog.Error("failed to get survey info csv", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey info csv"})
		return
	}
}

func (h *HttpEndpoints) getResponsesCount(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	filter, err := apihelpers.ParseFilterQueryFromCtx(c)
	if err != nil {
		slog.Error("failed to parse filter", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	filter["key"] = c.DefaultQuery("surveyKey", "")

	slog.Info("getting responses count", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	count, err := h.studyDBConn.GetResponsesCount(token.InstanceID, studyKey, filter)
	if err != nil {
		slog.Error("failed to get responses count", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get responses count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *HttpEndpoints) generateResponsesExport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getParticipantsCount(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	filter, err := apihelpers.ParseFilterQueryFromCtx(c)
	if err != nil {
		slog.Error("failed to parse filter", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("getting participants count", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	count, err := h.studyDBConn.GetParticipantCount(token.InstanceID, studyKey, filter)
	if err != nil {
		slog.Error("failed to get participants count", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get participants count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *HttpEndpoints) generateParticipantsExport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getReportsCount(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	filter, err := apihelpers.ParseFilterQueryFromCtx(c)
	if err != nil {
		slog.Error("failed to parse filter", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	reportKey := c.DefaultQuery("reportKey", "")
	if reportKey != "" {
		filter["key"] = reportKey
	}

	slog.Info("getting reports count", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	count, err := h.studyDBConn.GetReportCountForQuery(token.InstanceID, studyKey, filter)
	if err != nil {
		slog.Error("failed to get reports count", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reports count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *HttpEndpoints) generateReportsExport(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getConfidentialResponsesCount(c *gin.Context) {
	// TODO: implement
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteConfidentialResponse(c *gin.Context) {
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
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	query, err := apihelpers.ParseResponseExportQueryFromCtx(c)
	if err != nil || query == nil {
		slog.Error("failed to parse query", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	surveyKey := query.SurveyKey
	if surveyKey == "" {
		slog.Error("surveyKey is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "surveyKey is required"})
		return
	}

	query.PaginationInfos.Filter["key"] = surveyKey

	slog.Info("getting study responses", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))

	rawResponses, paginationInfo, err := h.studyDBConn.GetResponses(
		token.InstanceID,
		studyKey,
		query.PaginationInfos.Filter,
		query.PaginationInfos.Sort,
		query.PaginationInfos.Page,
		query.PaginationInfos.Limit,
	)
	if err != nil {
		slog.Error("failed to get study responses", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study responses"})
		return
	}

	surveyVersions, err := studydefinition.PrepareSurveyInfosFromDB(
		h.studyDBConn,
		token.InstanceID,
		studyKey,
		surveyKey,
		&surveydefinition.ExtractOptions{
			UseLabelLang: "",
			IncludeItems: nil,
			ExcludeItems: nil,
		},
	)
	if err != nil {
		slog.Error("failed to get survey versions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey versions"})
		return
	}

	respParser, err := surveyresponses.NewResponseParser(
		surveyKey,
		surveyVersions,
		query.UseShortKeys,
		query.IncludeMeta,
		query.QuestionOptionSep,
	)
	if err != nil {
		slog.Error("failed to create response parser", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create response parser"})
		return
	}

	responses := make([]map[string]interface{}, len(rawResponses))

	for i, rawResp := range rawResponses {
		resp, err := respParser.ParseResponse(&rawResp)
		if err != nil {
			slog.Error("failed to parse response", slog.String("error", err.Error()))
			continue
		}
		output, err := respParser.ResponseToFlatObj(resp)
		if err != nil {
			slog.Error("failed to convert response to flat object", slog.String("error", err.Error()))
			continue
		}
		responses[i] = output
	}

	c.JSON(http.StatusOK, gin.H{
		"responses":  responses,
		"pagination": paginationInfo,
	})
}

func (h *HttpEndpoints) getStudyResponseById(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	responseID := c.Param("responseID")

	slog.Info("getting study response by ID", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("responseID", responseID))

	query, err := apihelpers.ParseResponseExportQueryFromCtx(c)
	if err != nil {
		slog.Error("failed to parse response export query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	rawResponse, err := h.studyDBConn.GetResponseByID(token.InstanceID, studyKey, responseID)
	if err != nil {
		slog.Error("failed to get study response by ID", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study response by ID"})
		return
	}

	surveyVersions, err := studydefinition.PrepareSurveyInfosFromDB(
		h.studyDBConn,
		token.InstanceID,
		studyKey,
		rawResponse.Key,
		&surveydefinition.ExtractOptions{
			UseLabelLang: "",
			IncludeItems: nil,
			ExcludeItems: nil,
		},
	)
	if err != nil {
		slog.Error("failed to get survey versions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get survey versions"})
		return
	}

	respParser, err := surveyresponses.NewResponseParser(
		rawResponse.Key,
		surveyVersions,
		query.UseShortKeys,
		query.IncludeMeta,
		query.QuestionOptionSep,
	)
	if err != nil {
		slog.Error("failed to create response parser", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create response parser"})
		return
	}

	resp, err := respParser.ParseResponse(&rawResponse)
	if err != nil {
		slog.Error("failed to parse response", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	output, err := respParser.ResponseToFlatObj(resp)
	if err != nil {
		slog.Error("failed to convert response to flat object", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert response to flat object"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": output})
}

func (h *HttpEndpoints) deleteStudyResponses(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	query, err := apihelpers.ParseResponseExportQueryFromCtx(c)
	if err != nil {
		slog.Error("failed to parse response export query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	controlField := c.DefaultQuery("controlField", "")
	if controlField != studyKey {
		slog.Error("controlField does not match studyKey", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to delete study responses"})
		return
	}

	surveyKey := query.SurveyKey
	if surveyKey == "" {
		slog.Error("surveyKey is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "surveyKey is required"})
		return
	}

	filter := query.PaginationInfos.Filter
	filter["key"] = surveyKey // ensure surveyKey is included in the filter

	slog.Info("deleting study responses", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))

	err = h.studyDBConn.DeleteResponses(token.InstanceID, studyKey, filter)
	if err != nil {
		slog.Error("failed to delete study responses", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete study responses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study responses deleted"})
}

func (h *HttpEndpoints) deleteStudyResponse(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	responseID := c.Param("responseID")

	if responseID == "" {
		slog.Error("responseID is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "responseID is required"})
		return
	}

	slog.Info("deleting study response", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("responseID", responseID))

	err := h.studyDBConn.DeleteResponseByID(token.InstanceID, studyKey, responseID)
	if err != nil {
		slog.Error("failed to delete study response", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete study response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study response deleted"})
}

func (h *HttpEndpoints) getStudyParticipants(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")

	slog.Info("getting study participants", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	query, err := apihelpers.ParsePaginatedQueryFromCtx(c)
	if err != nil || query == nil {
		slog.Error("failed to parse paginated query", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	participants, paginationInfo, err := h.studyDBConn.GetParticipants(
		token.InstanceID,
		studyKey,
		query.Filter,
		query.Sort,
		query.Page,
		query.Limit,
	)
	if err != nil {
		slog.Error("failed to get study participants", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study participants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participants": participants,
		"pagination":   paginationInfo,
	})
}

func (h *HttpEndpoints) getStudyParticipant(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	participantID := c.Param("participantID")

	slog.Info("getting study participant", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("participantID", participantID))

	participant, err := h.studyDBConn.GetParticipantByID(token.InstanceID, studyKey, participantID)
	if err != nil {
		slog.Error("failed to get study participant", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"participant": participant})
}

func (h *HttpEndpoints) getStudyReports(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")

	slog.Info("getting study reports", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	query, err := apihelpers.ParsePaginatedQueryFromCtx(c)
	if err != nil || query == nil {
		slog.Error("failed to parse paginated query", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	reportKey := c.DefaultQuery("reportKey", "")
	if reportKey != "" {
		query.Filter["key"] = reportKey
	}

	reports, paginationInfo, err := h.studyDBConn.GetReports(
		token.InstanceID,
		studyKey,
		query.Filter,
		query.Page,
		query.Limit,
	)
	if err != nil {
		slog.Error("failed to get study reports", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports":    reports,
		"pagination": paginationInfo,
	})
}

func (h *HttpEndpoints) getStudyReport(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	reportID := c.Param("reportID")

	slog.Info("getting study report", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("reportID", reportID))

	report, err := h.studyDBConn.GetReportByID(token.InstanceID, studyKey, reportID)
	if err != nil {
		slog.Error("failed to get study report", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

func (h *HttpEndpoints) getStudyFiles(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	query, err := apihelpers.ParsePaginatedQueryFromCtx(c)
	if err != nil || query == nil {
		slog.Error("failed to parse query", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("getting study files", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	files, paginationInfo, err := h.studyDBConn.GetParticipantFileInfos(
		token.InstanceID,
		studyKey,
		query.Filter,
		query.Page,
		query.Limit,
	)
	if err != nil {
		slog.Error("failed to get study files", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fileInfos":  files,
		"pagination": paginationInfo,
	})
}

// download file
func (h *HttpEndpoints) getStudyFile(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	fileID := c.Param("fileID")

	slog.Info("getting study file", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("fileID", fileID))

	fileInfo, err := h.studyDBConn.GetParticipantFileInfoByID(token.InstanceID, studyKey, fileID)
	if err != nil {
		slog.Error("failed to get study file info", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study file info"})
		return
	}

	filePath := filepath.Join(h.filestorePath, fileInfo.Path)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Error("file does not exist", slog.String("path", filePath))
		c.JSON(http.StatusNotFound, gin.H{"error": "file does not exist"})
		return
	}

	// Return file from file system
	filenameToSave := filepath.Base(fileInfo.Path)
	c.Header("Content-Disposition", "attachment; filename="+filenameToSave)
	c.File(filePath)
}

func (h *HttpEndpoints) deleteStudyFile(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	fileID := c.Param("fileID")

	slog.Info("deleting study file", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("fileID", fileID))

	fileInfo, err := h.studyDBConn.GetParticipantFileInfoByID(token.InstanceID, studyKey, fileID)
	if err != nil {
		slog.Error("failed to get study file info", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get study file info"})
		return
	}

	// remove file from file system
	err = os.Remove(filepath.Join(h.filestorePath, fileInfo.Path))
	if err != nil {
		slog.Error("failed to delete study file", slog.String("error", err.Error()), slog.String("path", fileInfo.Path))
	}
	if fileInfo.PreviewPath != "" {
		err := os.Remove(filepath.Join(h.filestorePath, fileInfo.PreviewPath))
		if err != nil {
			slog.Error("failed to delete study file preview", slog.String("error", err.Error()), slog.String("path", fileInfo.PreviewPath))
		}
	}

	// delete file info from database
	err = h.studyDBConn.DeleteParticipantFileInfoByID(token.InstanceID, studyKey, fileID)
	if err != nil {
		slog.Error("failed to delete study file", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete study file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "study file deleted"})
}
