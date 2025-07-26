package apihandlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	pc "github.com/case-framework/case-backend/pkg/permission-checker"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func (h *HttpEndpoints) addParticipantManagementEndpoints(rg *gin.RouterGroup) {
	participantGroup := rg.Group("/participants")

	participantGroup.POST("/virtual", h.createVirtualParticipant)

	participantGroup.PUT("/:participantID", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType:        pc.RESOURCE_TYPE_STUDY,
			ResourceKeys:        []string{pc.RESOURCE_KEY_STUDY_ALL},
			ExtractResourceKeys: getStudyKeyFromParams,
			Action:              pc.ACTION_EDIT_PARTICIPANT_STATES,
		},
		nil,
		h.editStudyParticipant,
	))
}

func (h *HttpEndpoints) createVirtualParticipant(c *gin.Context) {

}

func (h *HttpEndpoints) editStudyParticipant(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	studyKey := c.Param("studyKey")
	participantID := c.Param("participantID")

	slog.Info("updating participant", slog.String("participantID", participantID), slog.String("studyKey", studyKey), slog.String("userID", token.Subject), slog.String("instanceID", token.InstanceID))

	var req studyTypes.Participant
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "participant updated",
		"participant": updatedParticipant,
	})
}
