package globalinfos

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var blockedJwtIndexNames []string

type BlockedJwt struct {
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

var indexesForBlockedJwtsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "token", Value: 1},
		},
		Options: options.Index().SetName("token_1"),
	},
	{
		Keys: bson.D{
			{Key: "expiresAt", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(0).SetName("expiresAt_1"),
	},
}

func (dbService *GlobalInfosDBService) DropIndexForBlockedJwtsCollection(dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		err := dbService.collectionBlockedJwts().Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for blocked jwts", slog.String("error", err.Error()))
		}
	} else {
		for _, indexName := range blockedJwtIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for blocked jwts collection")
				continue
			}
			err := dbService.collectionBlockedJwts().Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for blocked jwts", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *GlobalInfosDBService) CreateDefaultIndexesForBlockedJwtsCollection() {
	ctx, cancel := dbService.getContext()
	defer cancel()

	names, err := dbService.collectionBlockedJwts().Indexes().CreateMany(ctx, indexesForBlockedJwtsCollection)
	if err != nil {
		slog.Error("Error creating index for blocked jwts", slog.String("error", err.Error()))
	}
	blockedJwtIndexNames = names
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
		slog.Error("Error adding JWT to blocked list", slog.String("error", err.Error()))
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
		slog.Error("Error checking if JWT is blocked", slog.String("error", err.Error()))
		return false
	}
	return count > 0
}
