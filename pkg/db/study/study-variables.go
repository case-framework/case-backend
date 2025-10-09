package study

import (
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var indexesForStudyVariablesCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "studyKey", Value: 1},
			{Key: "key", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("studyKey_1_key_1"),
	},
}

func (dbService *StudyDBService) DropIndexForStudyVariablesCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyVariables(instanceID)
	if dropAll {
		_, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for studyVariables", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, index := range indexesForStudyVariablesCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for studyVariables collection", slog.String("index", fmt.Sprintf("%+v", index)), slog.String("instanceID", instanceID))
				continue
			}
			indexName := *index.Options.Name
			_, err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for studyVariables", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForStudyVariablesCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionStudyVariables(instanceID).Indexes().CreateMany(ctx, indexesForStudyVariablesCollection)
	if err != nil {
		slog.Error("Error creating index for studyVariables", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}
