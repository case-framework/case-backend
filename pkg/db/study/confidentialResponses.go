package study

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var indexesForConfidentialResponsesCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "participantID", Value: 1},
			{Key: "key", Value: 1},
		},
		Options: options.Index().SetName("participantID_key_1"),
	},
}

func (dbService *StudyDBService) DropIndexForConfidentialResponsesCollection(instanceID string, studyKey string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionConfidentialResponses(instanceID, studyKey)

	if dropAll {
		_, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for confidential responses", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
		}
	} else {
		for _, index := range indexesForConfidentialResponsesCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for confidential responses collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for confidential responses", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForConfidentialResponsesCollection(instanceID string, studyKey string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionConfidentialResponses(instanceID, studyKey)
	_, err := collection.Indexes().CreateMany(ctx, indexesForConfidentialResponsesCollection)
	if err != nil {
		slog.Error("Error creating index for confidential responses", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("studyKey", studyKey))
	}
}

func (dbService *StudyDBService) AddConfidentialResponse(instanceID string, studyKey string, response studyTypes.SurveyResponse) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	if len(response.ParticipantID) < 1 {
		return "", errors.New("participantID must be defined")
	}
	res, err := dbService.collectionConfidentialResponses(instanceID, studyKey).InsertOne(ctx, response)
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), err
}

func (dbService *StudyDBService) ReplaceConfidentialResponse(instanceID string, studyKey string, response studyTypes.SurveyResponse) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"participantID": response.ParticipantID,
		"key":           response.Key,
	}

	upsert := true
	options := options.ReplaceOptions{
		Upsert: &upsert,
	}
	_, err := dbService.collectionConfidentialResponses(instanceID, studyKey).ReplaceOne(ctx, filter, response, &options)
	return err
}

func (dbService *StudyDBService) FindConfidentialResponses(instanceID string, studyKey string, participantID string, key string) (responses []studyTypes.SurveyResponse, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if participantID == "" {
		return responses, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": participantID}
	if key != "" {
		filter["key"] = key
	}

	cur, err := dbService.collectionConfidentialResponses(instanceID, studyKey).Find(
		ctx,
		filter,
		nil,
	)

	if err != nil {
		return responses, err
	}
	defer cur.Close(ctx)

	responses = []studyTypes.SurveyResponse{}
	for cur.Next(ctx) {
		var result studyTypes.SurveyResponse
		err := cur.Decode(&result)
		if err != nil {
			return responses, err
		}

		responses = append(responses, result)
	}
	if err := cur.Err(); err != nil {
		return responses, err
	}

	return responses, nil
}

func (dbService *StudyDBService) FindAndExecuteOnConfidentialResponses(
	ctx context.Context,
	instanceID string, studyKey string,
	returnOnError bool,
	fn func(r studyTypes.SurveyResponse, args ...interface{}) error,
	args ...interface{},
) error {

	filter := bson.M{}
	cursor, err := dbService.collectionConfidentialResponses(instanceID, studyKey).Find(ctx, filter)
	if err != nil {
		return err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var response studyTypes.SurveyResponse
		if err = cursor.Decode(&response); err != nil {
			slog.Error("Error while decoding response", slog.String("error", err.Error()))
			continue
		}

		if err = fn(response, args...); err != nil {
			slog.Error("Error while executing function on confidential response", slog.String("responseID", response.ID.Hex()), slog.String("error", err.Error()))
			if returnOnError {
				return err
			}
			continue
		}
	}
	return nil
}

func (dbService *StudyDBService) DeleteConfidentialResponses(instanceID string, studyKey string, participantID string, key string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if participantID == "" {
		return 0, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": participantID}
	if key != "" {
		filter["key"] = key
	}

	res, err := dbService.collectionConfidentialResponses(instanceID, studyKey).DeleteMany(ctx, filter)
	return res.DeletedCount, err
}

func (dbService *StudyDBService) UpdateParticipantIDonConfidentialResponses(instanceID string, studyKey string, oldID string, newID string) (count int64, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if oldID == "" || newID == "" {
		return 0, errors.New("participant id must be defined")
	}
	filter := bson.M{"participantID": oldID}
	update := bson.M{"$set": bson.M{"participantID": newID}}

	res, err := dbService.collectionConfidentialResponses(instanceID, studyKey).UpdateMany(ctx, filter, update)
	return res.ModifiedCount, err
}
