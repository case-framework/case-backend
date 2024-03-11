package study

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	studytypes "github.com/case-framework/case-backend/pkg/types/study"
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

// get file infos by query and pagination
