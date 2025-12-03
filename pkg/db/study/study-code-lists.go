package study

import (
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studytypes "github.com/case-framework/case-backend/pkg/study/types"
)

var indexesForStudyCodeListsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "studyKey", Value: 1},
			{Key: "listKey", Value: 1},
			{Key: "code", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("studyKey_1_listKey_1_code_1"),
	},
}

func (dbService *StudyDBService) DropIndexForStudyCodeListsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyCodeLists(instanceID)
	if dropAll {
		_, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for studyCodeLists", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, index := range indexesForStudyCodeListsCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for studyCodeLists collection", slog.String("index", fmt.Sprintf("%+v", index)), slog.String("instanceID", instanceID))
				continue
			}
			indexName := *index.Options.Name
			_, err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for studyCodeLists", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForStudyCodeListsCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyCodeLists(instanceID)
	_, err := collection.Indexes().CreateMany(ctx, indexesForStudyCodeListsCollection)
	if err != nil {
		slog.Error("Error creating index for studyCodeLists", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
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

func (dbService *StudyDBService) GetStudyCodeListEntries(
	instanceID string,
	studyKey string,
	listKey string,
	page int64,
	limit int64,
) ([]studytypes.StudyCodeListEntry, *PaginationInfos, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
	}

	totalCount, err := dbService.collectionStudyCodeLists(instanceID).CountDocuments(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	paginationInfo := prepPaginationInfos(
		totalCount,
		page,
		limit,
	)

	skip := (paginationInfo.CurrentPage - 1) * paginationInfo.PageSize

	opts := options.Find().
		SetSkip(skip).
		SetLimit(paginationInfo.PageSize).
		SetSort(bson.D{
			{Key: "_id", Value: 1},
		})

	var entries []studytypes.StudyCodeListEntry
	cursor, err := dbService.collectionStudyCodeLists(instanceID).Find(ctx, filter, opts)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &entries)
	return entries, paginationInfo, err
}

func (dbService *StudyDBService) CountStudyCodeListEntries(instanceID string, studyKey string, listKey string) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
	}

	count, err := dbService.collectionStudyCodeLists(instanceID).CountDocuments(ctx, filter)
	return count, err
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

func (dbService *StudyDBService) DeleteStudyCodeListsForStudy(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionStudyCodeLists(instanceID).DeleteMany(ctx, bson.M{"studyKey": studyKey})
	return err
}

func (dbService *StudyDBService) DrawStudyCode(instanceID string, studyKey string, listKey string) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"studyKey": studyKey,
		"listKey":  listKey,
	}

	var result studytypes.StudyCodeListEntry
	err := dbService.collectionStudyCodeLists(instanceID).FindOneAndDelete(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	return result.Code, nil
}
