package apihelpers

import (
	"encoding/json"
	"strconv"

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

	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
	if err != nil {
		return nil, err
	}

	sort := bson.M{}
	if sortStr := c.DefaultQuery("sort", ""); sortStr != "" {
		if err := json.Unmarshal([]byte(sortStr), &sort); err != nil {
			return nil, err
		}
	}

	filter := bson.M{}
	if filterStr := c.DefaultQuery("filter", ""); filterStr != "" {
		if err := json.Unmarshal([]byte(filterStr), &filter); err != nil {
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
