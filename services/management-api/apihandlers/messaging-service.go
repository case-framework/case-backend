package apihandlers

import (
	"log/slog"
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	messagingDB "github.com/case-framework/case-backend/pkg/db/messaging"
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
	// TODO
	// TODO: check if templates are valid
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
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

func (h *HttpEndpoints) getStudyMessageTemplates(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) saveStudyMessageTemplate(c *gin.Context) {
	// TODO: check if templates are valid
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyMessageTemplate(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteStudyMessageTemplate(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
