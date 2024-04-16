package study

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func (dbService *StudyDBService) SaveResearcherMessage(instanceID string, studyKey string, message studyTypes.StudyMessage) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionResearcherMessages(instanceID, studyKey).InsertOne(ctx, message)
	return err
}

func (dbService *StudyDBService) FindResearcherMessages(instanceID string, studyKey string) (messages []studyTypes.StudyMessage, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}

	cur, err := dbService.collectionResearcherMessages(instanceID, studyKey).Find(
		ctx,
		filter,
	)
	if err != nil {
		return messages, err
	}
	defer cur.Close(ctx)

	messages = []studyTypes.StudyMessage{}
	for cur.Next(ctx) {
		var result studyTypes.StudyMessage
		err := cur.Decode(&result)
		if err != nil {
			return messages, err
		}

		messages = append(messages, result)
	}
	if err := cur.Err(); err != nil {
		return messages, err
	}

	return messages, nil
}

func (dbService *StudyDBService) DeleteResearcherMessages(instanceID string, studyKey string, messageIDs []string) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	idsToDelete := []primitive.ObjectID{}
	for _, id := range messageIDs {
		_id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			slog.Debug("unexpected error while converting id to ObjectID: %v", slog.String("error", err.Error()))
			continue
		}
		idsToDelete = append(idsToDelete, _id)
	}
	filter := bson.M{"_id": bson.M{"$in": idsToDelete}}

	res, err := dbService.collectionResearcherMessages(instanceID, studyKey).DeleteMany(ctx, filter)
	return res.DeletedCount, err
}
