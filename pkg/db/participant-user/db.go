package participantuser

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
	COLLECTION_NAME_PARTICIPANT_USERS   = "users"
	COLLECTION_NAME_RENEW_TOKENS        = "renewTokens"
	COLLECTION_NAME_OTPS                = "otps"
	COLLECTION_NAME_FAILED_OTP_ATTEMPTS = "failedOtpAttempts"
)

type ParticipantUserDBService struct {
	DBClient        *mongo.Client
	timeout         int
	noCursorTimeout bool
	DBNamePrefix    string
	InstanceIDs     []string
}

func NewParticipantUserDBService(configs db.DBConfig) (*ParticipantUserDBService, error) {
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

	puDBSc := &ParticipantUserDBService{
		DBClient:        dbClient,
		timeout:         configs.Timeout,
		noCursorTimeout: configs.NoCursorTimeout,
		DBNamePrefix:    configs.DBNamePrefix,
		InstanceIDs:     configs.InstanceIDs,
	}

	if configs.RunIndexCreation {
		puDBSc.ensureIndexes()
	}
	return puDBSc, nil
}

func (dbService *ParticipantUserDBService) getDBName(instanceID string) string {
	return dbService.DBNamePrefix + instanceID + "_users"
}

func (dbService *ParticipantUserDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

func (dbService *ParticipantUserDBService) collectionParticipantUsers(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_PARTICIPANT_USERS)
}

func (dbService *ParticipantUserDBService) collectionRenewTokens(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_RENEW_TOKENS)
}

func (dbService *ParticipantUserDBService) collectionOTPs(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_OTPS)
}

func (dbService *ParticipantUserDBService) collectionFailedOtpAttempts(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_FAILED_OTP_ATTEMPTS)
}

func (dbService *ParticipantUserDBService) ensureIndexes() {
	slog.Debug("Ensuring indexes for participant user DB")
	for _, instanceID := range dbService.InstanceIDs {

		err := dbService.CreateIndexForParticipantUsers(instanceID)
		if err != nil {
			slog.Debug("Error creating indexes for participant users: ", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		}

		err = dbService.CreateIndexForRenewTokens(instanceID)
		if err != nil {
			slog.Debug("Error creating indexes for renew tokens: ", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		}

		err = dbService.CreateIndexForOTPs(instanceID)
		if err != nil {
			slog.Debug("Error creating indexes for OTPs: ", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		}

		err = dbService.CreateIndexForFailedOtpAttempts(instanceID)
		if err != nil {
			slog.Debug("Error creating indexes for failed OTP attempts: ", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		}

		// Fix field name for contactInfos
		err = dbService.FixFieldNameForContactInfos(instanceID)
		if err != nil {
			slog.Debug("Error fixing field name for contactInfos: ", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
		}
	}
}
