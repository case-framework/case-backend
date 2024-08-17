package messaging

import (
	"context"
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collection names
const (
	COLLECTION_NAME_EMAIL_TEMPLATES = "email-templates"
	COLLECTION_NAME_SMS_TEMPLATES   = "sms-templates"
	COLLECTION_NAME_EMAIL_SCHEDULES = "auto-messages"
	COLLECTION_NAME_OUTGOING_EMAILS = "outgoing-emails"
	COLLECTION_NAME_SENT_EMAILS     = "sent-emails"
	COLLECTION_NAME_SENT_SMS        = "sent-sms"
)

type MessagingDBService struct {
	DBClient        *mongo.Client
	timeout         int
	noCursorTimeout bool
	DBNamePrefix    string
	InstanceIDs     []string
}

func NewMessagingDBService(configs db.DBConfig) (*MessagingDBService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	defer cancel()

	dbClient, err := mongo.Connect(ctx,
		options.Client().ApplyURI(configs.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(configs.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(configs.MaxPoolSize),
	)

	if err != nil {
		return nil, err
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()

	if err != nil {
		return nil, err
	}

	messagingDBSc := &MessagingDBService{
		DBClient:        dbClient,
		timeout:         configs.Timeout,
		noCursorTimeout: configs.NoCursorTimeout,
		DBNamePrefix:    configs.DBNamePrefix,
		InstanceIDs:     configs.InstanceIDs,
	}

	if configs.RunIndexCreation {
		if err := messagingDBSc.ensureIndexes(); err != nil {
			slog.Error("Error ensuring indexes for messaging DB: ", slog.String("error", err.Error()))
		}
	}

	return messagingDBSc, nil
}

func (dbService *MessagingDBService) getDBName(instanceID string) string {
	return dbService.DBNamePrefix + instanceID + "_messageDB"
}

func (dbService *MessagingDBService) collectionEmailTemplates(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_EMAIL_TEMPLATES)
}

func (dbService *MessagingDBService) collectionSMSTemplates(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_SMS_TEMPLATES)
}

func (dbService *MessagingDBService) collectionEmailSchedules(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_EMAIL_SCHEDULES)
}

func (dbService *MessagingDBService) collectionOutgoingEmails(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_OUTGOING_EMAILS)
}

func (dbService *MessagingDBService) collectionSentEmails(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_SENT_EMAILS)
}

func (dbService *MessagingDBService) collectionSentSMS(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_SENT_SMS)
}

func (dbService *MessagingDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

func (dbService *MessagingDBService) ensureIndexes() error {
	slog.Debug("Ensuring indexes for messaging DB")
	for _, instanceID := range dbService.InstanceIDs {
		ctx, cancel := dbService.getContext()
		defer cancel()

		// Email Templates
		_, err := dbService.collectionEmailTemplates(instanceID).Indexes().CreateOne(
			ctx,
			// index unique on messageType and studyKey combo:
			mongo.IndexModel{
				Keys: bson.D{
					{Key: "messageType", Value: 1},
					{Key: "studyKey", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		)
		if err != nil {
			slog.Error("Error creating index for email templates: ", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		}

		// Sent SMS
		err = dbService.CreateSentSMSIndex(instanceID)
		if err != nil {
			slog.Error("Error creating index for sent SMS: ", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		}

		// Outgoing Emails
		// add index generation here if needed

		// Sent Emails
		// add index generation here if needed

		// Email Schedules
		// add index generation here if needed
	}

	return nil
}
