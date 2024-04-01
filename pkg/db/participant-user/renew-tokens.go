package participantuser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

const (
	RENEW_TOKEN_GRACE_PERIOD     = 30 // seconds
	RENEW_TOKEN_DEFAULT_LIFETIME = 60 * 60 * 24 * 90
)

func (dbService *ParticipantUserDBService) CreateIndexForRenewTokens(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionRenewTokens(instanceID).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "userID", Value: 1},
					{Key: "renewToken", Value: 1},
					{Key: "expiresAt", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "expiresAt", Value: 1},
				},
				Options: options.Index().SetExpireAfterSeconds(RENEW_TOKEN_GRACE_PERIOD),
			},
			{
				Keys: bson.D{
					{Key: "renewToken", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	return err
}

func (dbService *ParticipantUserDBService) CreateRenewToken(instanceID string, userID string, token string, lifeTimeInSec int) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	ttl := time.Duration(lifeTimeInSec) * time.Second
	if lifeTimeInSec <= 0 {
		ttl = time.Duration(RENEW_TOKEN_DEFAULT_LIFETIME) * time.Second
	}
	renewToken := userTypes.RenewToken{
		UserID:     userID,
		RenewToken: token,
		ExpiresAt:  time.Now().Add(ttl),
	}

	_, err := dbService.collectionRenewTokens(instanceID).InsertOne(ctx, renewToken)
	return err
}
