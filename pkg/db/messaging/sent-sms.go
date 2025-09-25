package messaging

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/messaging/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var indexesForSentSMSCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userID", Value: 1},
			{Key: "sentAt", Value: 1},
			{Key: "messageType", Value: 1},
		},
		Options: options.Index().SetName("userID_sentAt_messageType_1"),
	},
}

func (dbService *MessagingDBService) DropIndexForSentSMSCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionSentSMS(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for sent SMS", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, index := range indexesForSentSMSCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for sent SMS collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionSentSMS(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for sent SMS", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *MessagingDBService) CreateDefaultIndexesForSentSMSCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionSentSMS(instanceID).Indexes().CreateMany(ctx, indexesForSentSMSCollection)
	if err != nil {
		slog.Error("Error creating index for sent SMS", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

func (dbService *MessagingDBService) AddToSentSMS(instanceID string, sms types.SentSMS) (types.SentSMS, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionSentSMS(instanceID).InsertOne(ctx, sms)
	if err != nil {
		return sms, err
	}
	sms.ID = res.InsertedID.(primitive.ObjectID)
	return sms, nil
}

func (dbService *MessagingDBService) CountSentSMSForUser(instanceID string, userID string, messageType string, sentAfter time.Time) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"userID": userID,
		"sentAt": bson.M{"$gt": sentAfter},
	}
	if messageType != "" {
		filter["messageType"] = messageType
	}

	return dbService.collectionSentSMS(instanceID).CountDocuments(ctx, filter)
}

func (dbService *MessagingDBService) GetAllSentSMSForUser(instanceID string, userID string, sentAfter time.Time) ([]types.SentSMS, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"userID": userID,
		"sentAt": bson.M{"$gt": sentAfter},
	}

	var sms []types.SentSMS
	cursor, err := dbService.collectionSentSMS(instanceID).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &sms); err != nil {
		return nil, err
	}
	return sms, nil
}
