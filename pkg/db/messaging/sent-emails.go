package messaging

import (
	"log/slog"
	"time"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const idxSentEmailsUserIDSentAt = "userId_1_sentAt_1"

var defaultSentEmailIndexNames = []string{
	idxSentEmailsUserIDSentAt,
}

var indexesForSentEmailsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "sentAt", Value: 1},
		},
		Options: options.Index().SetName(idxSentEmailsUserIDSentAt),
	},
}

func (dbService *MessagingDBService) DropIndexForSentEmailsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		err := dbService.collectionSentEmails(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for sent emails", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, indexName := range defaultSentEmailIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for sent emails collection")
				continue
			}
			err := dbService.collectionSentEmails(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for sent emails", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *MessagingDBService) CreateDefaultIndexesForSentEmailsCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionSentEmails(instanceID).Indexes().CreateMany(ctx, indexesForSentEmailsCollection)
	if err != nil {
		slog.Error("Error creating index for sent emails", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

func (dbService *MessagingDBService) AddToSentEmails(instanceID string, email messagingTypes.OutgoingEmail) (messagingTypes.OutgoingEmail, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	email.Content = ""
	email.SentAt = time.Now().UTC()
	email.To = []string{}

	email.ID = bson.NilObjectID
	res, err := dbService.collectionSentEmails(instanceID).InsertOne(ctx, email)
	if err != nil {
		return email, err
	}
	email.ID = res.InsertedID.(bson.ObjectID)
	return email, nil
}
