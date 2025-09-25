package messaging

import (
	"fmt"
	"log/slog"
	"time"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var indexesForSentEmailsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "sentAt", Value: 1},
		},
		Options: options.Index().SetName("userId_sentAt_1"),
	},
}

func (dbService *MessagingDBService) DropIndexForSentEmailsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionSentEmails(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for sent emails", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, index := range indexesForSentEmailsCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for sent emails collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionSentEmails(instanceID).Indexes().DropOne(ctx, indexName)
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

	email.ID = primitive.NilObjectID
	res, err := dbService.collectionSentEmails(instanceID).InsertOne(ctx, email)
	if err != nil {
		return email, err
	}
	email.ID = res.InsertedID.(primitive.ObjectID)
	return email, nil
}
