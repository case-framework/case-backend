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

		// TODO: study permissions
	}
}

func getStudyKeyFromParams(c *gin.Context) []string {
	return []string{c.Param("studyKey")}
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
