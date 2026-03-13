package messaging

import (
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/messaging/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var sentSMSIndexNames []string

var indexesForSentSMSCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userID", Value: 1},
			{Key: "sentAt", Value: 1},
			{Key: "messageType", Value: 1},
		},
		Options: options.Index().SetName("userID_1_sentAt_1_messageType_1"),
	},
}

func (dbService *MessagingDBService) DropIndexForSentSMSCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		err := dbService.collectionSentSMS(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for sent SMS", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, indexName := range sentSMSIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for sent SMS collection")
				continue
			}
			err := dbService.collectionSentSMS(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for sent SMS", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *MessagingDBService) CreateDefaultIndexesForSentSMSCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	names, err := dbService.collectionSentSMS(instanceID).Indexes().CreateMany(ctx, indexesForSentSMSCollection)
	if err != nil {
		slog.Error("Error creating index for sent SMS", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
	sentSMSIndexNames = names
}

func (dbService *MessagingDBService) AddToSentSMS(instanceID string, sms types.SentSMS) (types.SentSMS, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionSentSMS(instanceID).InsertOne(ctx, sms)
	if err != nil {
		return sms, err
	}
	sms.ID = res.InsertedID.(bson.ObjectID)
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
