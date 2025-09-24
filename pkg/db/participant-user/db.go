package participantuser

import (
	"context"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collection names
const (
	COLLECTION_NAME_PARTICIPANT_USERS           = "users"
	COLLECTION_NAME_PARTICIPANT_USER_ATTRIBUTES = "userAttributes"
	COLLECTION_NAME_RENEW_TOKENS                = "renewTokens"
	COLLECTION_NAME_OTPS                        = "otps"
	COLLECTION_NAME_FAILED_OTP_ATTEMPTS         = "failedOtpAttempts"
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

func (dbService *ParticipantUserDBService) collectionParticipantUserAttributes(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_PARTICIPANT_USER_ATTRIBUTES)
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

func (dbService *ParticipantUserDBService) CreateDefaultIndexes() {
	for _, instanceID := range dbService.InstanceIDs {
		dbService.CreateDefaultIndexesForParticipantUsersCollection(instanceID)
		dbService.CreateDefaultIndexesForParticipantUserAttributesCollection(instanceID)
		dbService.CreateDefaultIndexesForRenewTokensCollection(instanceID)
		dbService.CreateDefaultIndexesForOTPsCollection(instanceID)
		dbService.CreateDefaultIndexesForFailedOtpAttemptsCollection(instanceID)
	}
}

func (dbService *ParticipantUserDBService) DropIndexes(dropAll bool) {
	for _, instanceID := range dbService.InstanceIDs {
		dbService.DropIndexForParticipantUsersCollection(instanceID, dropAll)
		dbService.DropIndexForParticipantUserAttributesCollection(instanceID, dropAll)
		dbService.DropIndexForRenewTokensCollection(instanceID, dropAll)
		dbService.DropIndexForOTPsCollection(instanceID, dropAll)
		dbService.DropIndexForFailedOtpAttemptsCollection(instanceID, dropAll)
	}
}
