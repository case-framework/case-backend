package participantuser

import (
	"errors"
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

func (dbService *ParticipantUserDBService) CreateIndexForRenewTokens(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if _, err := dbService.collectionRenewTokens(instanceID).Indexes().DropAll(ctx); err != nil {
		slog.Error("Error dropping indexes for renew tokens", slog.String("error", err.Error()))
	}

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

func (dbService *ParticipantUserDBService) DeleteRenewTokenByToken(instanceID string, token string) error {
	filter := bson.M{"renewToken": token}

	ctx, cancel := dbService.getContext()
	defer cancel()
	res, err := dbService.collectionRenewTokens(instanceID).DeleteOne(ctx, filter, nil)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("no renew token oject found with the given token value")
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
