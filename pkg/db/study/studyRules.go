package study

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (dbService *StudyDBService) CreateIndexForStudyRulesCollection(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "studyKey", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "uploadedAt", Value: 1},
				{Key: "studyKey", Value: 1},
			},
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (dbService *StudyDBService) deleteStudyRules(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)
	_, err := collection.DeleteMany(ctx, bson.M{"studyKey": studyKey})
	return err
}
