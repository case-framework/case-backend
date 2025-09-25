package db

import (
	"context"
	"errors"

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
