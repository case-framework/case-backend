package study

import (
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studytypes "github.com/case-framework/case-backend/pkg/study/types"
)

var indexesForParticipantFilesCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "participantID", Value: 1},
			{Key: "createdAt", Value: -1},
		},
		Options: options.Index().SetName("participantID_1_createdAt_-1"),
	},
	{
		Keys: bson.D{
			{Key: "createdAt", Value: -1},
		},
		Options: options.Index().SetName("createdAt_-1"),
	},
}

func (dbService *StudyDBService) DropIndexForParticipantFilesCollection(instanceID string, studyKey string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionFiles(instanceID, studyKey)

	if dropAll {
		_, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for participant files", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		}
	} else {
		for _, index := range indexesForParticipantFilesCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for participant files collection", slog.String("index", fmt.Sprintf("%+v", index)), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
				continue
			}
			indexName := *index.Options.Name
			_, err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for participant files", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForParticipantFilesCollection(instanceID string, studyKey string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionFiles(instanceID, studyKey)
	_, err := collection.Indexes().CreateMany(ctx, indexesForParticipantFilesCollection)
	if err != nil {
		slog.Error("Error creating index for participant files", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
	}
}

// get one by id
func (dbService *StudyDBService) GetParticipantFileInfoByID(instanceID string, studyKey string, fileInfoID string) (participantFileInfo studytypes.FileInfo, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(fileInfoID)
	if err != nil {
		return participantFileInfo, err
	}

	filter := bson.M{
		"_id": _id,
	}

	err = dbService.collectionFiles(instanceID, studyKey).FindOne(ctx, filter).Decode(&participantFileInfo)
	return participantFileInfo, err
}

// delete one by id
func (dbService *StudyDBService) DeleteParticipantFileInfoByID(instanceID string, studyKey string, fileInfoID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(fileInfoID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}

	res, err := dbService.collectionFiles(instanceID, studyKey).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return err
}

// count by query
func (dbService *StudyDBService) CountParticipantFileInfos(instanceID string, studyKey string, query bson.M) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	return dbService.collectionFiles(instanceID, studyKey).CountDocuments(ctx, query)
}

// get file infos by query and pagination
var sortBySubmittedAt = bson.D{
	primitive.E{Key: "submittedAt", Value: -1},
}

func (dbService *StudyDBService) GetParticipantFileInfos(instanceID string, studyKey string, query bson.M, page int64, limit int64) (fileInfos []studytypes.FileInfo, paginationInfo *PaginationInfos, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	count, err := dbService.CountParticipantFileInfos(instanceID, studyKey, query)
	if err != nil {
		return fileInfos, paginationInfo, err
	}

	paginationInfo = prepPaginationInfos(
		count,
		page,
		limit,
	)

	opts := options.Find()
	opts.SetSort(sortBySubmittedAt)
	skip := (paginationInfo.CurrentPage - 1) * paginationInfo.PageSize
	opts.SetSkip(skip)
	opts.SetLimit(paginationInfo.PageSize)

	cursor, err := dbService.collectionFiles(instanceID, studyKey).Find(ctx, query, opts)
	if err != nil {
		return fileInfos, paginationInfo, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &fileInfos)
	return fileInfos, paginationInfo, err
}

// save file info
func (dbService *StudyDBService) CreateParticipantFileInfo(instanceID string, studyKey string, fileInfo studytypes.FileInfo) (studytypes.FileInfo, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	result, err := dbService.collectionFiles(instanceID, studyKey).InsertOne(ctx, fileInfo)
	if err != nil {
		return fileInfo, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		fileInfo.ID = oid
	}

	return fileInfo, nil
}

// update file info path and status
func (dbService *StudyDBService) UpdateParticipantFileInfoPathAndStatus(instanceID string, studyKey string, fileInfoID string, path string, status string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(fileInfoID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}
	update := bson.M{
		"$set": bson.M{
			"path":      path,
			"status":    status,
			"updatedAt": time.Now(),
		},
	}
	_, err = dbService.collectionFiles(instanceID, studyKey).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}
