package participantuser

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	OTP_TTL = 60 * 5 // 5 minutes
)

func (dbService *ParticipantUserDBService) CreateIndexForOTPs(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionOTPs(instanceID).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "userID", Value: 1},
					{Key: "code", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: "createdAt", Value: 1},
				},
				Options: options.Index().SetExpireAfterSeconds(OTP_TTL),
			},
		},
	)
	return err
}
