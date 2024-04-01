package participantuser

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
