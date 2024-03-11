package apihelpers

import (
	"encoding/json"
	"net/url"
	"strconv"

	surveyresponses "github.com/case-framework/case-backend/pkg/exporter/survey-responses"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type PagenatedQuery struct {
	Page   int64
	Limit  int64
	Sort   bson.M
	Filter bson.M
}

func ParsePaginatedQueryFromCtx(c *gin.Context) (*PagenatedQuery, error) {
	page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}

	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	if err != nil {
		return nil, err
	}

	if limit < 1 {
		limit = 10
	}

	sort := bson.M{}
	if sortStr := c.DefaultQuery("sort", ""); sortStr != "" {
		decodedSortStr, err := url.QueryUnescape(sortStr)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(decodedSortStr), &sort); err != nil {
			return nil, err
		}
	}

	filter := bson.M{}
	if filterStr := c.DefaultQuery("filter", ""); filterStr != "" {
		decodedFilterStr, err := url.QueryUnescape(filterStr)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(decodedFilterStr), &filter); err != nil {
			return nil, err
		}
	}

	return &PagenatedQuery{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Filter: filter,
	}, nil
}

type ResponseExportQuery struct {
	SurveyKey         string
	UseShortKeys      bool
	QuestionOptionSep string
	IncludeMeta       *surveyresponses.IncludeMeta
	PaginationInfos   *PagenatedQuery
}

func ParseResponseExportQueryFromCtx(c *gin.Context) (*ResponseExportQuery, error) {
	paginatedQuery, err := ParsePaginatedQueryFromCtx(c)
	if err != nil {
		return nil, err
	}

	surveyKey := c.DefaultQuery("surveyKey", "")
	useShortKeys, err := strconv.ParseBool(c.DefaultQuery("shortKeys", "false"))
	if err != nil {
		return nil, err
	}

	questionOptionSep := c.DefaultQuery("questionOptionSep", "-")

	includeMeta := &surveyresponses.IncludeMeta{}
	// TODO

	return &ResponseExportQuery{
		SurveyKey:         surveyKey,
		UseShortKeys:      useShortKeys,
		QuestionOptionSep: questionOptionSep,
		IncludeMeta:       includeMeta,
		PaginationInfos:   paginatedQuery,
	}, nil
}
