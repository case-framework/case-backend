package apihelpers

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	surveyresponses "github.com/case-framework/case-backend/pkg/study/exporter/survey-responses"
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

	sort, err := ParseSortQueryFromCtx(c)
	if err != nil {
		return nil, err
	}

	filter, err := ParseFilterQueryFromCtx(c)
	if err != nil {
		return nil, err
	}

	return &PagenatedQuery{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Filter: filter,
	}, nil
}

func ParseFilterQueryFromCtx(c *gin.Context) (bson.M, error) {
	return ParseEscapedJSONQueryFromContext(c, "filter")
}

func ParseSortQueryFromCtx(c *gin.Context) (bson.M, error) {
	return ParseEscapedJSONQueryFromContext(c, "sort")
}

func ParseEscapedJSONQueryFromContext(c *gin.Context, key string) (bson.M, error) {
	jsonMap := bson.M{}
	jsonStr := c.DefaultQuery(key, "")
	if jsonStr == "" {
		return jsonMap, nil
	}

	decodedJSONStr, err := url.QueryUnescape(jsonStr)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(decodedJSONStr), &jsonMap); err != nil {
		return nil, err
	}

	return jsonMap, nil
}

type ResponseExportQuery struct {
	SurveyKey         string
	UseShortKeys      bool
	QuestionOptionSep string
	Format            string
	IncludeMeta       *surveyresponses.IncludeMeta
	PaginationInfos   *PagenatedQuery
	ExtraCtxCols      *[]string
}

func ParseResponseExportQueryFromCtx(c *gin.Context) (*ResponseExportQuery, error) {
	paginatedQuery, err := ParsePaginatedQueryFromCtx(c)
	if err != nil {
		return nil, err
	}

	surveyKey := c.DefaultQuery("surveyKey", "")
	if paginatedQuery == nil {
		paginatedQuery = &PagenatedQuery{
			Filter: bson.M{},
		}
	}
	paginatedQuery.Filter["key"] = surveyKey
	useShortKeys, err := strconv.ParseBool(c.DefaultQuery("shortKeys", "false"))
	if err != nil {
		return nil, err
	}

	questionOptionSep := c.DefaultQuery("questionOptionSep", "-")

	format := c.DefaultQuery("format", "wide")
	q := &ResponseExportQuery{
		SurveyKey:         surveyKey,
		UseShortKeys:      useShortKeys,
		QuestionOptionSep: questionOptionSep,
		Format:            format,
		PaginationInfos:   paginatedQuery,
	}

	extraCtxColsQuery := c.DefaultQuery("extraContextColumns", "")
	if extraCtxColsQuery != "" {
		*q.ExtraCtxCols = strings.Split(extraCtxColsQuery, ",")
	}

	// TODO
	includeMeta := &surveyresponses.IncludeMeta{}
	q.IncludeMeta = includeMeta

	return q, nil
}
