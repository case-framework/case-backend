package study

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

// get studies
func (dbService *StudyDBService) GetStudies(instanceID string, statusFilter string, onlyKeys bool) (studies []studyTypes.Study, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{}
	if statusFilter != "" {
		filter["status"] = statusFilter
	}
	opts := options.Find()
	if onlyKeys {
		projection := bson.D{
			primitive.E{Key: "key", Value: 1},
			primitive.E{Key: "secretKey", Value: 1},
			primitive.E{Key: "configs.idMappingMethod", Value: 1},
		}
		opts.SetProjection(projection)
	}
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &studies)
	if err != nil {
		return nil, err
	}

	return studies, nil
}
