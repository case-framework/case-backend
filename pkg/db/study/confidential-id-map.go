package study

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var confidentialIDMapIndexNames []string

var indexesForConfidentialIDMapCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "confidentialID", Value: 1},
			{Key: "studyKey", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("confidentialID_1_studyKey_1"),
	},
}

func (dbService *StudyDBService) DropIndexForConfidentialIDMapCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionConfidentialIDMap(instanceID)
	if dropAll {
		err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for confidentialIDMap", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, indexName := range confidentialIDMapIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for confidentialIDMap collection")
				continue
			}
			err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for confidentialIDMap", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForConfidentialIDMapCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	names, err := dbService.collectionConfidentialIDMap(instanceID).Indexes().CreateMany(ctx, indexesForConfidentialIDMapCollection)
	if err != nil {
		slog.Error("Error creating index for confidentialIDMap", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
	confidentialIDMapIndexNames = names
}

func (dbService *StudyDBService) AddConfidentialIDMapEntry(instanceID, confidentialID, profileID, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	entry := bson.M{
		"confidentialID": confidentialID,
		"profileID":      profileID,
		"studyKey":       studyKey,
	}

	_, err := dbService.collectionConfidentialIDMap(instanceID).InsertOne(ctx, entry)
	return err
}

func (dbService *StudyDBService) GetProfileIDFromConfidentialID(instanceID, confidentialID, studyKey string) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"confidentialID": confidentialID,
		"studyKey":       studyKey,
	}

	var result struct {
		ProfileID string `bson:"profileID"`
	}
	err := dbService.collectionConfidentialIDMap(instanceID).FindOne(ctx, filter).Decode(&result)
	return result.ProfileID, err
}

func (dbService *StudyDBService) RemoveConfidentialIDMapEntriesForStudy(instanceID, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionConfidentialIDMap(instanceID).DeleteMany(ctx, bson.M{"studyKey": studyKey})
	return err
}

func (dbService *StudyDBService) RemoveConfidentialIDMapEntriesForProfile(instanceID, profileID, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionConfidentialIDMap(instanceID).DeleteMany(ctx, bson.M{"profileID": profileID, "studyKey": studyKey})
	return err
}
