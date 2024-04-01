package globalinfos

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *GlobalInfosDBService) CreateIndexForTemptokens() error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionTemptokens().Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "userID", Value: 1},
					{Key: "instanceID", Value: 1},
					{Key: "purpose", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "expiration", Value: 1},
				},
				Options: options.Index().SetExpireAfterSeconds(0),
			},
			{
				Keys: bson.D{
					{Key: "token", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	return err
}
