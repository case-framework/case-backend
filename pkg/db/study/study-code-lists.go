package study

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/case-framework/case-backend/pkg/study/types"
	studytypes "github.com/case-framework/case-backend/pkg/study/types"
)

func (dbService *StudyDBService) CreateIndexForStudyCodeListsCollection(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if _, err := dbService.collectionStudyCodeLists(instanceID).Indexes().DropAll(ctx); err != nil {
		slog.Error("Error dropping indexes for studyCodeLists", slog.String("error", err.Error()))
	}

	collection := dbService.collectionStudyCodeLists(instanceID)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "studyKey", Value: 1},
				{Key: "listKey", Value: 1},
				{Key: "code", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (dbService *StudyDBService) AddStudyCodeListEntry(instanceID string, studyKey string, listKey string, code string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	entry := studytypes.StudyCodeListEntry{
		StudyKey: studyKey,
		ListKey:  listKey,
		Code:     code,
		AddedAt:  time.Now(),
	}

	_, err := dbService.collectionStudyCodeLists(instanceID).InsertOne(ctx, entry)
	return err
}

func (dbService *StudyDBService) GetUniqueStudyCodeListKeysForStudy(instanceID string, studyKey string) ([]string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var listKeys []string
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"studyKey": studyKey}}},
		{{Key: "$group", Value: bson.M{"_id": "$listKey"}}},
	}

	cursor, err := dbService.collectionStudyCodeLists(instanceID).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var results []struct {
		ListKey string `bson:"_id"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, r := range results {
		listKeys = append(listKeys, r.ListKey)
	}
	return listKeys, nil
}

func (dbService *StudyDBService) GetStudyCodeListEntries(instanceID string, studyKey string, listKey string) ([]studytypes.StudyCodeListEntry, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
	}

	var entries []studytypes.StudyCodeListEntry
	cursor, err := dbService.collectionStudyCodeLists(instanceID).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &entries)
	return entries, err
}

func (dbService *StudyDBService) StudyCodeListEntryExists(instanceID string, studyKey string, listKey string, code string) (bool, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
		"code":     code,
	}

	count, err := dbService.collectionStudyCodeLists(instanceID).CountDocuments(ctx, filter)
	return count > 0, err
}

func (dbService *StudyDBService) DeleteStudyCodeListEntries(instanceID string, studyKey string, listKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
	}

	_, err := dbService.collectionStudyCodeLists(instanceID).DeleteMany(ctx, filter)
	return err
}

func (dbService *StudyDBService) DeleteStudyCodeListEntry(instanceID string, studyKey string, listKey string, code string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
		"code":     code,
	}

	_, err := dbService.collectionStudyCodeLists(instanceID).DeleteOne(ctx, filter)
	return err
}

func (dbService *StudyDBService) DrawStudyCode(instanceID string, studyKey string, listKey string) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
	}

	var result types.StudyCodeListEntry
	err := dbService.collectionStudyCodeLists(instanceID).FindOneAndDelete(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	return result.Code, nil
}
