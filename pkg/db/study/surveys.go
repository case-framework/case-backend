package study

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *StudyDBService) CreateIndexForSurveyCollection(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionSurveys(instanceID, studyKey)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "surveyDefinition.key", Value: 1},
				{Key: "unpublished", Value: 1},
				{Key: "published", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "published", Value: 1},
				{Key: "surveyDefinition.key", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "unpublished", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "surveyDefinition.key", Value: 1},
				{Key: "versionID", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (dbService *StudyDBService) GetSurveyKeysForStudy(instanceID string, studyKey string, includeUnpublished bool) (surveyKeys []string, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if !includeUnpublished {
		filter["unpublished"] = 0
	}
	res, err := dbService.collectionSurveys(instanceID, studyKey).Distinct(ctx, "surveyDefinition.key", filter)
	if err != nil {
		return surveyKeys, err
	}
	surveyKeys = make([]string, len(res))
	for i, r := range res {
		surveyKeys[i] = r.(string)
	}
	return surveyKeys, err
}
