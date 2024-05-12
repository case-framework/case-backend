package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"

	studyService "github.com/case-framework/case-backend/pkg/study"
)

func (h *HttpEndpoints) AddStudyServiceAPI(rg *gin.RouterGroup) {
	studyServiceGroup := rg.Group("/study-service")

	/* studiesGroup := studyServiceGroup.Group("/studies")
	{
		studiesGroup.GET("/", h.getAllStudies) // ?status=active
		studiesGroup.GET("/:studyKey", h.getStudy)
		studiesGroup.GET("/participating", mw.GetAndValidateParticipantUserJWT(h.tokenSignKey), h.getParticipatingStudies)
	} */

	// study events
	eventsGroup := studyServiceGroup.Group("/events/:studyKey")
	eventsGroup.Use(mw.GetAndValidateParticipantUserJWT(h.tokenSignKey))
	{
		eventsGroup.POST("/enter", h.enterStudy)
		eventsGroup.POST("/custom", h.customStudyEvent)
		//eventsGroup.POST("/submit", h.submitSurvey)
		// eventsGroup.POST("/leave", h.leaveStudy)
	}

	// get assigned surveys
	// get survey infos
	// get survey for participant

	// merge temporary participant

	// delete files
	// file upload
}

func (h *HttpEndpoints) enterStudy(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	studyKey := c.Param("studyKey")

	var req struct {
		ProfileID string `json:"profileID"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.checkProfileBelongsToUser(token.InstanceID, token.Subject, req.ProfileID) {
		slog.Warn("profile not found", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("profileID", req.ProfileID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
		return
	}

	slog.Info("entering study", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	result, err := studyService.OnEnterStudy(token.InstanceID, studyKey, req.ProfileID)
	if err != nil {
		slog.Error("error entering study", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error entering study"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignedSurveys": result})
}

func (h *HttpEndpoints) customStudyEvent(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	studyKey := c.Param("studyKey")

	var req struct {
		EventKey  string                 `json:"eventKey"`
		ProfileID string                 `json:"profileId"`
		Payload   map[string]interface{} `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.checkProfileBelongsToUser(token.InstanceID, token.Subject, req.ProfileID) {
		slog.Warn("profile not found", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("profileID", req.ProfileID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
		return
	}

	if req.EventKey == "" {
		slog.Error("eventKey is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventKey is required"})
		return
	}

	slog.Info("custom study event", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("eventKey", req.EventKey))

	result, err := studyService.OnCustomStudyEvent(token.InstanceID, studyKey, req.ProfileID, req.EventKey, req.Payload)
	if err != nil {
		slog.Error("error entering study", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error entering study"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignedSurveys": result})
}
