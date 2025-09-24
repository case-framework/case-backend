package globalinfos

import (
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	umUtils "github.com/case-framework/case-backend/pkg/user-management/utils"
)

var indexesForTemptokensCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userID", Value: 1},
			{Key: "instanceID", Value: 1},
			{Key: "purpose", Value: 1},
		},
		Options: options.Index().SetName("userID_instanceID_purpose_1"),
	},
	{
		Keys: bson.D{
			{Key: "expiration", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(0).SetName("expiration_1"),
	},
	{
		Keys: bson.D{
			{Key: "token", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("token_1"),
	},
}

func (dbService *GlobalInfosDBService) DropIndexForTemptokensCollection(dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		if _, err := dbService.collectionTemptokens().Indexes().DropAll(ctx); err != nil {
			slog.Error("Error dropping indexes for temptokens", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForTemptokensCollection {
			indexName := *index.Options.Name
			_, err := dbService.collectionTemptokens().Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for temptokens", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *GlobalInfosDBService) CreateDefaultIndexesForTemptokensCollection() {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionTemptokens().Indexes().CreateMany(ctx, indexesForTemptokensCollection)
	if err != nil {
		slog.Error("Error creating index for temptokens", slog.String("error", err.Error()))
	}
}

func (dbService *GlobalInfosDBService) AddTempToken(t userTypes.TempToken) (token string, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	t.Token, err = umUtils.GenerateUniqueTokenString()
	if err != nil {
		return token, err
	}

	_, err = dbService.collectionTemptokens().InsertOne(ctx, t)
	if err != nil {
		return token, err
	}
	token = t.Token
	return
}

func (dbService *GlobalInfosDBService) DeleteAllTempTokenForUser(instanceID string, userID string, purpose string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"instanceID": instanceID, "userID": userID}
	if len(purpose) > 0 {
		filter["purpose"] = purpose
	}
	_, err := dbService.collectionTemptokens().DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func (dbService *GlobalInfosDBService) GetTempToken(token string) (userTypes.TempToken, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"token": token}

	t := userTypes.TempToken{}
	err := dbService.collectionTemptokens().FindOne(ctx, filter).Decode(&t)
	return t, err
}

func (dbService *GlobalInfosDBService) DeleteTempToken(token string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"token": token}
	res, err := dbService.collectionTemptokens().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("document not found")
	}
	return nil
}

func (dbService *GlobalInfosDBService) UpdateTempTokenExpirationTime(token string, newExpiration time.Time) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"token": token}

	update := bson.M{"$set": bson.M{"expiration": newExpiration}}
	_, err := dbService.collectionTemptokens().UpdateOne(ctx, filter, update)
	return err
}
