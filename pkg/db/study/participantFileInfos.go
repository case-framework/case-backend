package study

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// count

// get one by id

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
