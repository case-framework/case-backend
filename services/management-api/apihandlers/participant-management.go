package apihandlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	pc "github.com/case-framework/case-backend/pkg/permission-checker"
	studyService "github.com/case-framework/case-backend/pkg/study"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func (h *HttpEndpoints) addParticipantManagementEndpoints(rg *gin.RouterGroup) {
	participantGroup := rg.Group("/participants")

	participantGroup.POST("/virtual", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_CREATE_VIRTUAL_PARTICIPANT,
		},
		nil,
		h.createVirtualParticipant,
	))

	participantGroup.POST("/:participantID/responses", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_EDIT_PARTICIPANT_DATA,
		},
		nil,
		h.submitParticipantResponse,
	))

	participantGroup.POST("/:participantID/events", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_EDIT_PARTICIPANT_DATA,
		},
		nil,
		h.submitParticipantEvent,
	))

	participantGroup.POST("/:participantID/reports", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_EDIT_PARTICIPANT_DATA,
		},
		nil,
		h.submitParticipantReport,
	))

	/*
		Merge participants
		POST /v1/studies/:studyKey/participants/merge/
	*/

	participantGroup.PUT("/:participantID", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_EDIT_PARTICIPANT_DATA,
		},
		nil,
		h.editStudyParticipant,
	))
}

func (h *HttpEndpoints) createVirtualParticipant(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")

	slog.Info("creating virtual participant", slog.String("studyKey", studyKey), slog.String("userID", token.Subject), slog.String("instanceID", token.InstanceID))

	pState, err := studyService.OnRegisterVirtualParticipant(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("failed to create virtual participant", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create virtual participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "virtual participant created",
		"participant": pState,
	})
}

func (h *HttpEndpoints) submitParticipantResponse(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	participantID := c.Param("participantID")

	var req studyTypes.SurveyResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("submitting response for participant", slog.String("participantID", participantID), slog.String("studyKey", studyKey), slog.String("userID", token.Subject), slog.String("instanceID", token.InstanceID),
		slog.String("forSurveyKey", req.Key),
	)

	result, err := studyService.OnSubmitResponseOnBehalfOfParticipant(token.InstanceID, studyKey, participantID, req, token.Subject)
	if err != nil {
		slog.Error("failed to submit response", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "response submitted", "result": result})
}

type ParticipantEventRequest struct {
	EventKey string         `json:"eventKey"`
	Payload  map[string]any `json:"payload"`
}

func (h *HttpEndpoints) submitParticipantEvent(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	participantID := c.Param("participantID")

	var req ParticipantEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := studyService.OnCustomStudyEventOnBehalfOfParticipant(token.InstanceID, studyKey, participantID, req.EventKey, req.Payload)
	if err != nil {
		slog.Error("failed to submit event", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event submitted", "result": result})
}

func (h *HttpEndpoints) submitParticipantReport(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	participantID := c.Param("participantID")

	var report studyTypes.Report
	if err := c.ShouldBindJSON(&report); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if report.ParticipantID != participantID {
		slog.Error("participant ID in request does not match participant ID in path", slog.String("requestParticipantID", report.ParticipantID), slog.String("pathParticipantID", participantID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	slog.Info("submitting report for participant", slog.String("participantID", participantID), slog.String("studyKey", studyKey), slog.String("userID", token.Subject), slog.String("instanceID", token.InstanceID), slog.String("reportKey", report.Key))

	err := h.studyDBConn.SaveReport(token.InstanceID, studyKey, report)
	if err != nil {
		slog.Error("failed to save report", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "report submitted"})
}

func (h *HttpEndpoints) editStudyParticipant(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	participantID := c.Param("participantID")

	slog.Info("updating participant", slog.String("participantID", participantID), slog.String("studyKey", studyKey), slog.String("userID", token.Subject), slog.String("instanceID", token.InstanceID))

	var req studyTypes.Participant
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ParticipantID != participantID {
		slog.Error("participant ID in request does not match participant ID in path", slog.String("requestParticipantID", req.ParticipantID), slog.String("pathParticipantID", participantID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updatedParticipant, err := h.studyDBConn.UpdateParticipantIfNotModified(token.InstanceID, studyKey, req)
	if err != nil {
		slog.Error("failed to update participant", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "participant updated",
		"participant": updatedParticipant,
	})
}
