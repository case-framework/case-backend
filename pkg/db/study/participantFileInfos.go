package study

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studytypes "github.com/case-framework/case-backend/pkg/study/study"
)

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
