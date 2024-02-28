package apihandlers

import (
	"net/http"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
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
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) saveGlobalMessageTemplate(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
	// TODO: check if templates are valid

func (h *HttpEndpoints) getGlobalMessageTemplate(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) deleteGlobalMessageTemplate(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) getStudyMessageTemplates(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *HttpEndpoints) saveStudyMessageTemplate(c *gin.Context) {
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
	// TODO: check if templates are valid
