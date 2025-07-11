package globalinfos

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlockedJwt struct {
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

func (dbService *GlobalInfosDBService) CreateIndexForBlockedJwts() error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if _, err := dbService.collectionBlockedJwts().Indexes().DropAll(ctx); err != nil {
		slog.Error("Error dropping indexes for blocked jwts", slog.String("error", err.Error()))
	}

	_, err := dbService.collectionBlockedJwts().Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "token", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "expiresAt", Value: 1},
				},
				Options: options.Index().SetExpireAfterSeconds(0),
			},
		},
	)
	return err
}

// AddBlockedJwt adds a JWT token to the blocked list with the specified expiration time
func (dbService *GlobalInfosDBService) AddBlockedJwt(token string, expiresAt time.Time) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	blockedJwt := BlockedJwt{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	_, err := dbService.collectionBlockedJwts().InsertOne(ctx, blockedJwt)
	if err != nil {
		slog.Error("Error adding JWT to blocked list", slog.String("error", err.Error()), slog.String("token", token))
		return err
	}

	return nil
}

// IsJwtBlocked checks if a JWT token is in the blocked list
// This method uses the indexed token field for efficient lookup
func (dbService *GlobalInfosDBService) IsJwtBlocked(token string) bool {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"token": token}

	// Use CountDocuments with limit 1 for a faster existence check
	count, err := dbService.collectionBlockedJwts().CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		slog.Error("Error checking if JWT is blocked", slog.String("error", err.Error()), slog.String("token", token))
		return false
	}
	return count > 0
}
