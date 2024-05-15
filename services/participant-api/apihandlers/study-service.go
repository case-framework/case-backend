package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"

	studyService "github.com/case-framework/case-backend/pkg/study"
	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
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
		// eventsGroup.POST("/submit", h.submitSurvey)
		// eventsGroup.POST("/leave", h.leaveStudy)
		// merge temporary participant
	}

	participantInfoGroup := studyServiceGroup.Group("/participant-data/:studyKey")
	participantInfoGroup.Use(mw.GetAndValidateParticipantUserJWT(h.tokenSignKey))
	{
		participantInfoGroup.GET("/surveys", h.getAssignedSurveys)
		participantInfoGroup.GET("/survey/:surveyKey", h.getSurveyWithContext) // ?pid=profileID

		// delete files
		// file upload

		// reports:
		// get reports reports/studyKey - query for profileIDs, report key, page, limit, filter
	}

	// temporary participants
	tempParticipantGroup := studyServiceGroup.Group("/temp-participant")
	{
		tempParticipantGroup.POST("/register", mw.RequirePayload(), h.registerTempParticipant)
		// tempParticipantGroup.GET("/surveys", h.getTempParticipantSurveys)                     // ?pid=profileID&instanceID=instanceID&studyKey=studyKey
		tempParticipantGroup.GET("/survey", h.getTempParticipantSurveyWithContext) // ?pid=profileID&instanceID=instanceID&studyKey=studyKey&surveyKey=surveyKey
		tempParticipantGroup.POST("/submit-response", mw.RequirePayload(), h.submitTempParticipantResponse)
	}
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

	if req.ProfileID == "" {
		slog.Error("profileID is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "profileID is required"})
		return
	}

	if !h.checkProfileBelongsToUser(token.InstanceID, token.Subject, req.ProfileID) {
		slog.Warn("profile not found", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("profileID", req.ProfileID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
		return
	}

	slog.Debug("entering study", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

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
		ProfileID string                 `json:"profileID"`
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

	if req.ProfileID == "" {
		slog.Error("profileID is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "profileID is required"})
		return
	}

	slog.Debug("custom study event", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("eventKey", req.EventKey))

	result, err := studyService.OnCustomStudyEvent(token.InstanceID, studyKey, req.ProfileID, req.EventKey, req.Payload)
	if err != nil {
		slog.Error("error entering study", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error entering study"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignedSurveys": result})
}

func (h *HttpEndpoints) getAssignedSurveys(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getSurveyWithContext(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ParticipantUserClaims)

	studyKey := c.Param("studyKey")
	surveyKey := c.Param("surveyKey")
	pid := c.DefaultQuery("pid", "")

	if pid == "" {
		slog.Error("profileID is required", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "profileID is required"})
		return
	}

	if !h.checkProfileBelongsToUser(token.InstanceID, token.Subject, pid) {
		slog.Warn("profile not found", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("profileID", pid))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "profile not found"})
		return
	}

	slog.Info("getting survey with context", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey), slog.String("profileID", pid))

	result, err := studyService.GetAssignedSurveyWithContext(token.InstanceID, studyKey, surveyKey, pid)
	if err != nil {
		slog.Error("error getting survey with context", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting survey with context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"surveyWithContext": result})
}

func (h *HttpEndpoints) registerTempParticipant(c *gin.Context) {
	var req struct {
		InstanceID string `json:"instanceId"`
		StudyKey   string `json:"studyKey"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.InstanceID == "" || req.StudyKey == "" {
		slog.Error("missing required fields", slog.String("instanceID", req.InstanceID), slog.String("studyKey", req.StudyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	if !h.isInstanceAllowed(req.InstanceID) {
		slog.Error("instance not allowed", slog.String("instanceID", req.InstanceID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "instance not allowed"})
		return
	}

	slog.Info("registering temporary participant", slog.String("instanceID", req.InstanceID), slog.String("studyKey", req.StudyKey))

	pState, err := studyService.OnRegisterTempParticipant(req.InstanceID, req.StudyKey)
	if err != nil {
		slog.Error("error registering temporary participant", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error registering temporary participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"participant": pState})
}

func (h *HttpEndpoints) getTempParticipantSurveyWithContext(c *gin.Context) {
	instanceID := c.DefaultQuery("instanceID", "")
	studyKey := c.DefaultQuery("studyKey", "")
	surveyKey := c.DefaultQuery("surveyKey", "")
	pid := c.DefaultQuery("pid", "")

	if !h.isInstanceAllowed(instanceID) {
		slog.Error("instance not allowed", slog.String("instanceID", instanceID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "instance not allowed"})
		return
	}

	if instanceID == "" || studyKey == "" || surveyKey == "" {
		slog.Error("missing required fields", slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("surveyKey", surveyKey))
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	result, err := studyService.GetSurveyWithContextForTempParticipant(instanceID, studyKey, surveyKey, pid)
	if err != nil {
		slog.Error("error getting survey with context", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting survey with context"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"surveyWithContext": result})
}

func (h *HttpEndpoints) submitTempParticipantResponse(c *gin.Context) {
	var req struct {
		InstanceID string                    `json:"instanceId"`
		StudyKey   string                    `json:"studyKey"`
		Pid        string                    `json:"pid"`
		Response   studyTypes.SurveyResponse `json:"response"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.InstanceID == "" || req.StudyKey == "" || req.Pid == "" {
		slog.Error("missing required fields", slog.String("instanceID", req.InstanceID), slog.String("studyKey", req.StudyKey), slog.String("pid", req.Pid))
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	if !h.isInstanceAllowed(req.InstanceID) {
		slog.Error("instance not allowed", slog.String("instanceID", req.InstanceID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "instance not allowed"})
		return
	}

	slog.Info("submitting response for temporary participant", slog.String("instanceID", req.InstanceID), slog.String("studyKey", req.StudyKey), slog.String("pid", req.Pid))

	result, err := studyService.OnSubmitResponseForTempParticipant(req.InstanceID, req.StudyKey, req.Pid, req.Response)
	if err != nil {
		slog.Error("error submitting response for temporary participant", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error submitting response for temporary participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignedSurveys": result})
}
