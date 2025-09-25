package participantuser

import (
	"errors"
	"fmt"
	"log/slog"
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

var indexesForRenewTokensCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userID", Value: 1},
			{Key: "renewToken", Value: 1},
			{Key: "expiresAt", Value: 1},
		},
		Options: options.Index().SetName("userID_1_renewToken_1_expiresAt_1"),
	},
	{
		Keys: bson.D{
			{Key: "userID", Value: 1},
			{Key: "sessionID", Value: 1},
		},
		Options: options.Index().SetName("userID_1_sessionID_1"),
	},
	{
		Keys: bson.D{
			{Key: "expiresAt", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(RENEW_TOKEN_GRACE_PERIOD).SetName("expiresAt_1"),
	},
	{
		Keys: bson.D{
			{Key: "renewToken", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("uniq_renewToken_1"),
	},
}

func (dbService *ParticipantUserDBService) DropIndexForRenewTokensCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionRenewTokens(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for renew tokens", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForRenewTokensCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for renew tokens collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionRenewTokens(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for renew tokens", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ParticipantUserDBService) CreateDefaultIndexesForRenewTokensCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionRenewTokens(instanceID).Indexes().CreateMany(
		ctx, indexesForRenewTokensCollection,
	)
	if err != nil {
		slog.Error("Error creating index for renew tokens", slog.String("error", err.Error()))
	}
}

func (dbService *ParticipantUserDBService) CreateRenewToken(instanceID string, userID string, token string, lifeTimeInSec int, sessionID string) error {
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
		SessionID:  sessionID,
	}

	_, err := dbService.collectionRenewTokens(instanceID).InsertOne(ctx, renewToken)
	return err
}

func (dbService *ParticipantUserDBService) DeleteRenewTokenByToken(instanceID string, token string) error {
	filter := bson.M{"renewToken": token}

	ctx, cancel := dbService.getContext()
	defer cancel()
	res, err := dbService.collectionRenewTokens(instanceID).DeleteOne(ctx, filter, nil)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("no renew token object found with the given token value")
	}
	return nil
}

func (dbService *ParticipantUserDBService) DeleteRenewTokensForUser(instanceID string, userID string) (int64, error) {
	filter := bson.M{"userID": userID}

	ctx, cancel := dbService.getContext()
	defer cancel()
	res, err := dbService.collectionRenewTokens(instanceID).DeleteMany(ctx, filter, nil)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (dbService *ParticipantUserDBService) DeleteRenewTokensForSession(instanceID string, userID string, sessionID string) (int64, error) {
	filter := bson.M{"userID": userID, "sessionID": sessionID}

	ctx, cancel := dbService.getContext()
	defer cancel()
	res, err := dbService.collectionRenewTokens(instanceID).DeleteMany(ctx, filter, nil)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (dbService *ParticipantUserDBService) FindAndUpdateRenewToken(instanceID string, userID string, renewToken string, nextToken string) (rtObj userTypes.RenewToken, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userID": userID, "renewToken": renewToken, "expiresAt": bson.M{"$gt": time.Now()}}
	updatePipeline := bson.A{
		bson.M{
			"$set": bson.M{
				"nextToken": bson.M{
					"$cond": bson.A{
						bson.M{
							"$eq": bson.A{
								bson.M{"$ifNull": bson.A{"$nextToken", nil}},
								nil,
							},
						},
						nextToken,
						"$nextToken",
					},
				},
				"expiresAt": bson.M{
					"$cond": bson.A{
						bson.M{
							"$eq": bson.A{
								bson.M{"$ifNull": bson.A{"$nextToken", nil}},
								nil,
							},
						},
						time.Now().Add(RENEW_TOKEN_GRACE_PERIOD * time.Second),
						"$expiresAt",
					},
				},
			},
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err = dbService.collectionRenewTokens(instanceID).FindOneAndUpdate(ctx, filter, updatePipeline, opts).Decode(&rtObj)
	return
}
