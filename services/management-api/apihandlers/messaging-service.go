package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
	emailtemplates "github.com/case-framework/case-backend/pkg/email-templates"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	"github.com/gin-gonic/gin"

	pc "github.com/case-framework/case-backend/pkg/permission-checker"
)

func (h *HttpEndpoints) AddMessagingServiceAPI(rg *gin.RouterGroup) {
	messagingGroup := rg.Group("/messaging")

	messagingGroup.Use(mw.GetAndValidateManagementUserJWT(h.tokenSignKey))
	messagingGroup.Use(mw.IsInstanceIDInJWTAllowed(h.allowedInstanceIDs))

	emailTemplatesGroup := messagingGroup.Group("/email-templates")

	// Global email templates
	h.addMessagingGlobalEmailTemplatesAPI(emailTemplatesGroup)

	// Add study email templates
	h.addMessagingStudyEmailTemplatesAPI(emailTemplatesGroup)
}

func (h *HttpEndpoints) addMessagingGlobalEmailTemplatesAPI(rg *gin.RouterGroup) {

	rg.GET("/global-templates", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_GLOBAL_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.getGlobalMessageTemplates,
	))

	rg.POST("/global-templates", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_GLOBAL_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.saveGlobalMessageTemplate,
	))
	rg.GET("/global-templates/:messageType", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_GLOBAL_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.getGlobalMessageTemplate,
	))

	rg.DELETE("/global-templates/:messageType", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_GLOBAL_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.deleteGlobalMessageTemplate,
	))
}

func (h *HttpEndpoints) addMessagingStudyEmailTemplatesAPI(rg *gin.RouterGroup) {
	rg.GET("/study-templates", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_STUDY_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.getStudyMessageTemplatesForAllStudies,
	))
	rg.GET("/study-templates/:studyKey", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_STUDY_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		getStudyKeyLimiterFromContext,
		h.getStudyMessageTemplates,
	))
	rg.POST("/study-templates/:studyKey", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_STUDY_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		getStudyKeyLimiterFromContext,
		h.saveStudyMessageTemplate,
	))
	rg.GET("/study-templates/:studyKey/:messageType", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_STUDY_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		getStudyKeyLimiterFromContext,
		h.getStudyMessageTemplate,
	))
	rg.DELETE("/study-templates/:studyKey/:messageType", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_STUDY_EMAIL_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		getStudyKeyLimiterFromContext,
		h.deleteStudyMessageTemplate,
	))
}

func getStudyKeyLimiterFromContext(c *gin.Context) map[string]string {
	return map[string]string{"studyKey": c.Param("studyKey")}
}

func (h *HttpEndpoints) getGlobalMessageTemplates(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	slog.Info("getGlobalMessageTemplates: getting global message templates", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	messages, err := h.messagingDBConn.GetEmailTemplatesForAllStudies(token.InstanceID)
	if err != nil {
		slog.Error("getGlobalMessageTemplates: error getting global message templates", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting global message templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": messages})
}

func (h *HttpEndpoints) saveGlobalMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	// parse body
	var template messagingDB.EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		slog.Error("saveGlobalMessageTemplate: error parsing request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request body"})
		return
	}

	err := emailtemplates.CheckAllTranslationsParsable(template)
	if err != nil {
		slog.Error("saveGlobalMessageTemplate: error parsing template", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while checking template validity"})
		return
	}

	slog.Info("saveGlobalMessageTemplate: saving global message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	savedTemplate, err := h.messagingDBConn.SaveEmailTemplate(token.InstanceID, template)
	if err != nil {
		slog.Error("saveGlobalMessageTemplate: error saving global message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving global message template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": savedTemplate})
}

func (h *HttpEndpoints) getGlobalMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	messageType := c.Param("messageType")

	slog.Info("getGlobalMessageTemplate: getting global message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("messageType", messageType))

	message, err := h.messagingDBConn.GetGlobalEmailTemplateByMessageType(token.InstanceID, messageType)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			dummyTemplate := messagingDB.EmailTemplate{
				MessageType:  messageType,
				Translations: []messagingDB.LocalizedTemplate{},
			}
			c.JSON(http.StatusOK, gin.H{"template": dummyTemplate})
			return
		}

		slog.Error("getGlobalMessageTemplate: error getting global message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting global message template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": message})
}

func (h *HttpEndpoints) deleteGlobalMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	messageType := c.Param("messageType")

	slog.Info("deleteGlobalMessageTemplate: deleting global message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("messageType", messageType))

	err := h.messagingDBConn.DeleteEmailTemplate(token.InstanceID, messageType, "")
	if err != nil {
		slog.Error("deleteGlobalMessageTemplate: error deleting global message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting global message template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

func (h *HttpEndpoints) getStudyMessageTemplatesForAllStudies(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	slog.Info("getStudyMessageTemplatesForAllStudies: getting study message templates", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	messages, err := h.messagingDBConn.GetEmailTemplatesForAllStudies(token.InstanceID)
	if err != nil {
		slog.Error("getStudyMessageTemplatesForAllStudies: error getting study message templates", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting study message templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": messages})
}

func (h *HttpEndpoints) getStudyMessageTemplates(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")

	slog.Info("getStudyMessageTemplates: getting study message templates", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	messages, err := h.messagingDBConn.GetStudyEmailTemplates(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("getStudyMessageTemplates: error getting study message templates", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting study message templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": messages})
}

func (h *HttpEndpoints) saveStudyMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")

	// parse body
	var template messagingDB.EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		slog.Error("saveStudyMessageTemplate: error parsing request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request body"})
		return
	}
	template.StudyKey = studyKey

	err := emailtemplates.CheckAllTranslationsParsable(template)
	if err != nil {
		slog.Error("saveStudyMessageTemplate: error parsing template", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while checking template validity"})
		return
	}

	slog.Info("saveStudyMessageTemplate: saving study message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	savedTemplate, err := h.messagingDBConn.SaveEmailTemplate(token.InstanceID, template)
	if err != nil {
		slog.Error("saveStudyMessageTemplate: error saving study message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving study message template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": savedTemplate})
}

func (h *HttpEndpoints) getStudyMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")
	messageType := c.Param("messageType")

	slog.Info("getStudyMessageTemplate: getting study message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("messageType", messageType))

	message, err := h.messagingDBConn.GetStudyEmailTemplateByMessageType(token.InstanceID, studyKey, messageType)
	if err != nil {
		slog.Error("getStudyMessageTemplate: error getting study message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting study message template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": message})
}

func (h *HttpEndpoints) deleteStudyMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")
	messageType := c.Param("messageType")

	slog.Info("deleteStudyMessageTemplate: deleting study message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("messageType", messageType))

	err := h.messagingDBConn.DeleteEmailTemplate(token.InstanceID, messageType, studyKey)
	if err != nil {
		slog.Error("deleteStudyMessageTemplate: error deleting study message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting study message template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}
