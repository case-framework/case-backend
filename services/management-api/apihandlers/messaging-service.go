package apihandlers

import (
	"log/slog"
	"net/http"
	"time"

	mw "github.com/case-framework/case-backend/pkg/apihelpers/middlewares"
	jwthandling "github.com/case-framework/case-backend/pkg/jwt-handling"
	emailtemplates "github.com/case-framework/case-backend/pkg/messaging/email-templates"
	"github.com/case-framework/case-backend/pkg/messaging/templates"
	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
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

	// Scheduled emails
	scheduledEmailsGroup := messagingGroup.Group("/scheduled-emails")
	h.addMessagingScheduledEmailsAPI(scheduledEmailsGroup)

	// SMS templates
	smsTemplatesGroup := messagingGroup.Group("/sms-templates")
	h.addMessagingSMSTemplatesAPI(smsTemplatesGroup)
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

func (h *HttpEndpoints) addMessagingSMSTemplatesAPI(rg *gin.RouterGroup) {
	smsTemplatesGroup := rg.Group("/")

	smsTemplatesGroup.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_SMS_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.saveSMSTemplate,
	))
	smsTemplatesGroup.GET("/:messageType", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_SMS_TEMPLATES},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.getSMSTemplate,
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

func (h *HttpEndpoints) addMessagingScheduledEmailsAPI(rg *gin.RouterGroup) {
	rg.GET("/", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_SCHEDULED_EMAILS},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.getScheduledEmails,
	))

	rg.POST("/", mw.RequirePayload(), h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_SCHEDULED_EMAILS},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.SaveScheduledEmail,
	))

	rg.GET("/:id", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_SCHEDULED_EMAILS},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.getScheduledEmail,
	))

	rg.DELETE("/:id", h.useAuthorisedHandler(
		RequiredPermission{
			ResourceType: pc.RESOURCE_TYPE_MESSAGING,
			ResourceKeys: []string{pc.RESOURCE_KEY_MESSAGING_SCHEDULED_EMAILS},
			Action:       pc.ACTION_ALL,
		},
		nil,
		h.deleteScheduledEmail,
	))
}

func (h *HttpEndpoints) getGlobalMessageTemplates(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	slog.Info("getting global message templates", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	messages, err := h.messagingDBConn.GetGlobalEmailTemplates(token.InstanceID)
	if err != nil {
		slog.Error("error getting global message templates", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting global message templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": messages})
}

func (h *HttpEndpoints) saveGlobalMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	// parse body
	var template messagingTypes.EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		slog.Error("error parsing request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request body"})
		return
	}

	err := emailtemplates.CheckAllTranslationsParsable(template)
	if err != nil {
		slog.Error("error parsing template", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while checking template validity"})
		return
	}

	slog.Info("saving global message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	savedTemplate, err := h.messagingDBConn.SaveEmailTemplate(token.InstanceID, template)
	if err != nil {
		slog.Error("error saving global message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving global message template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": savedTemplate})
}

func (h *HttpEndpoints) getGlobalMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	messageType := c.Param("messageType")

	slog.Info("getting global message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("messageType", messageType))

	message, err := h.messagingDBConn.GetGlobalEmailTemplateByMessageType(token.InstanceID, messageType)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			dummyTemplate := messagingTypes.EmailTemplate{
				MessageType:  messageType,
				Translations: []messagingTypes.LocalizedTemplate{},
			}
			c.JSON(http.StatusOK, gin.H{"template": dummyTemplate})
			return
		}

		slog.Error("error getting global message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting global message template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": message})
}

func (h *HttpEndpoints) deleteGlobalMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	messageType := c.Param("messageType")

	slog.Info("deleting global message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("messageType", messageType))

	err := h.messagingDBConn.DeleteEmailTemplate(token.InstanceID, messageType, "")
	if err != nil {
		slog.Error("error deleting global message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting global message template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

func (h *HttpEndpoints) getSMSTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	messageType := c.Param("messageType")

	slog.Info("getting SMS template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("messageType", messageType))

	message, err := h.messagingDBConn.GetGlobalEmailTemplateByMessageType(token.InstanceID, messageType)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			dummyTemplate := messagingTypes.SMSTemplate{
				MessageType:  messageType,
				Translations: []messagingTypes.LocalizedTemplate{},
			}
			c.JSON(http.StatusOK, gin.H{"template": dummyTemplate})
			return
		}

		slog.Error("error getting SMS template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting SMS template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": message})
}

func (h *HttpEndpoints) saveSMSTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	// parse body
	var template messagingTypes.SMSTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		slog.Error("error parsing request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request body"})
		return
	}

	err := templates.CheckAllTranslationsParsable(template.Translations, template.MessageType)
	if err != nil {
		slog.Error("error parsing template", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while checking template validity"})
		return
	}

	slog.Info("saving SMS template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	savedTemplate, err := h.messagingDBConn.SaveSMSTemplate(token.InstanceID, template)
	if err != nil {
		slog.Error("error saving SMS template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving SMS template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": savedTemplate})
}

func (h *HttpEndpoints) getStudyMessageTemplatesForAllStudies(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	slog.Info("getting study message templates", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	messages, err := h.messagingDBConn.GetEmailTemplatesForAllStudies(token.InstanceID)
	if err != nil {
		slog.Error("error getting study message templates", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting study message templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": messages})
}

func (h *HttpEndpoints) getStudyMessageTemplates(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")

	slog.Info("getting study message templates", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	messages, err := h.messagingDBConn.GetStudyEmailTemplates(token.InstanceID, studyKey)
	if err != nil {
		slog.Error("error getting study message templates", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting study message templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"templates": messages})
}

func (h *HttpEndpoints) saveStudyMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")

	// parse body
	var template messagingTypes.EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		slog.Error("error parsing request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request body"})
		return
	}
	template.StudyKey = studyKey

	err := emailtemplates.CheckAllTranslationsParsable(template)
	if err != nil {
		slog.Error("error parsing template", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while checking template validity"})
		return
	}

	slog.Info("saving study message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey))

	savedTemplate, err := h.messagingDBConn.SaveEmailTemplate(token.InstanceID, template)
	if err != nil {
		slog.Error("error saving study message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving study message template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": savedTemplate})
}

func (h *HttpEndpoints) getStudyMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")
	messageType := c.Param("messageType")

	slog.Info("getting study message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("messageType", messageType))

	message, err := h.messagingDBConn.GetStudyEmailTemplateByMessageType(token.InstanceID, studyKey, messageType)
	if err != nil {
		slog.Error("error getting study message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting study message template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"template": message})
}

func (h *HttpEndpoints) deleteStudyMessageTemplate(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	studyKey := c.Param("studyKey")
	messageType := c.Param("messageType")

	slog.Info("deleting study message template", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("studyKey", studyKey), slog.String("messageType", messageType))

	err := h.messagingDBConn.DeleteEmailTemplate(token.InstanceID, messageType, studyKey)
	if err != nil {
		slog.Error("error deleting study message template", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting study message template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

func (h *HttpEndpoints) getScheduledEmails(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	slog.Info("getting scheduled emails", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	schedules, err := h.messagingDBConn.GetAllScheduledEmails(token.InstanceID)
	if err != nil {
		slog.Error("error getting scheduled emails", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting scheduled emails"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

func (h *HttpEndpoints) SaveScheduledEmail(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)

	// parse body
	var schedule messagingTypes.ScheduledEmail
	if err := c.ShouldBindJSON(&schedule); err != nil {
		slog.Error("error parsing request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error parsing request body"})
		return
	}

	slog.Info("saving scheduled email", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject))

	// check if template is valid
	err := emailtemplates.CheckAllTranslationsParsable(schedule.Template)
	if err != nil {
		slog.Error("error parsing template", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while checking template validity"})
		return
	}

	// ensure that times are in the future
	if 0 < schedule.Until {
		if schedule.Until < time.Now().Unix() {
			slog.Error("error saving scheduled email", slog.String("error", "invalid termination date of auto message schedule, is in past"))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid termination date of auto message schedule, is in past"})
			return
		}
		if schedule.Until < schedule.NextTime {
			slog.Error("error saving scheduled email", slog.String("error", "invalid termination date of auto message schedule, earlier than start date"))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid termination date of auto message schedule, earlier than start date"})
			return
		}
	}

	savedSchedule, err := h.messagingDBConn.SaveScheduledEmail(token.InstanceID, schedule)
	if err != nil {
		slog.Error("error saving scheduled email", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error saving scheduled email"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schedule": savedSchedule})
}

func (h *HttpEndpoints) getScheduledEmail(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	id := c.Param("id")

	slog.Info("getting scheduled email", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("id", id))

	schedule, err := h.messagingDBConn.GetScheduledEmailByID(token.InstanceID, id)
	if err != nil {
		slog.Error("error getting scheduled email", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting scheduled email"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func (h *HttpEndpoints) deleteScheduledEmail(c *gin.Context) {
	token := c.MustGet("validatedToken").(*jwthandling.ManagementUserClaims)
	id := c.Param("id")

	slog.Info("deleting scheduled email", slog.String("instanceID", token.InstanceID), slog.String("userID", token.Subject), slog.String("id", id))

	err := h.messagingDBConn.DeleteScheduledEmail(token.InstanceID, id)
	if err != nil {
		slog.Error("error deleting scheduled email", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting scheduled email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule deleted"})
}
