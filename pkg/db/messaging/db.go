package messaging

import (
	"context"
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
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

func (dbService *MessagingDBService) CreateDefaultIndexes() {
	for _, instanceID := range dbService.InstanceIDs {
		start := time.Now()
		slog.Info("Creating default indexes for messaging DB", slog.String("instanceID", instanceID))
		dbService.CreateDefaultIndexesForEmailTemplatesCollection(instanceID)
		dbService.CreateDefaultIndexesForSMSTemplatesCollection(instanceID)
		// email schedules collection has no default indexes at the moment
		// outgoing emails collection has no default indexes at the moment
		dbService.CreateDefaultIndexesForSentEmailsCollection(instanceID)
		dbService.CreateDefaultIndexesForSentSMSCollection(instanceID)
		slog.Info("Default indexes created for messaging DB", slog.String("instanceID", instanceID), slog.String("duration", time.Since(start).String()))
	}
}

func (dbService *MessagingDBService) DropIndexes(dropAll bool) {
	for _, instanceID := range dbService.InstanceIDs {
		start := time.Now()
		slog.Info("Dropping indexes for messaging DB", slog.String("instanceID", instanceID))
		dbService.DropIndexForEmailTemplatesCollection(instanceID, dropAll)
		dbService.DropIndexForSMSTemplatesCollection(instanceID, dropAll)
		dbService.DropIndexForEmailSchedulesCollection(instanceID, dropAll)
		dbService.DropIndexForOutgoingEmailsCollection(instanceID, dropAll)
		dbService.DropIndexForSentEmailsCollection(instanceID, dropAll)
		dbService.DropIndexForSentSMSCollection(instanceID, dropAll)
		slog.Info("Indexes dropped for messaging DB", slog.String("instanceID", instanceID), slog.String("duration", time.Since(start).String()))
	}
}
