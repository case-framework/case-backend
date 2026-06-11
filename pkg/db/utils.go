package db

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ListCollectionIndexes(ctx context.Context, collection *mongo.Collection) ([]bson.M, error) {
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		var cmdErr mongo.CommandError
		if errors.As(err, &cmdErr) && cmdErr.Code == 26 {
			return []bson.M{}, nil
		}
		return nil, err
	}
	defer cursor.Close(ctx)

	indexes := []bson.M{}
	if err = cursor.All(ctx, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

func ConnectWithRetry[T any](dbLabel string, maxAttempts int, retryDelay time.Duration, connect func() (T, error)) (T, error) {
	var result T
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err = connect()
		if err == nil {
			if attempt > 1 {
				slog.Info("Connected to DB after retry", slog.String("db", dbLabel), slog.Int("attempt", attempt))
			}
			return result, nil
		}

		slog.Error("Error connecting to DB", slog.String("db", dbLabel), slog.Int("attempt", attempt), slog.Int("maxAttempts", maxAttempts), slog.String("error", err.Error()))
		if attempt < maxAttempts {
			time.Sleep(retryDelay)
		}
	}

	return result, err
}
