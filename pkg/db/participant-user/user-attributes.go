package participantuser

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

var indexesForParticipantUserAttributesCollection = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "type", Value: 1}},
		Options: options.Index().SetName("idx_user_attributes_userId").SetUnique(true),
	},
}

func (dbService *ParticipantUserDBService) DropIndexForParticipantUserAttributesCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for participant user attributes", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForParticipantUserAttributesCollection {
			if index.Options.Name == nil {
				slog.Error("Index name is nil for participant user attributes collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for participant user attributes", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ParticipantUserDBService) CreateDefaultIndexesForParticipantUserAttributesCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().CreateMany(ctx, indexesForParticipantUserAttributesCollection)
	if err != nil {
		slog.Error("Error creating index for participant user attributes", slog.String("error", err.Error()))
	}
}

// Create or update a user attribute for a user by type
func (dbService *ParticipantUserDBService) SetUserAttribute(
	instanceID string,
	userID string,
	attributeType string,
	attributes map[string]any,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	userIDObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = dbService.collectionParticipantUserAttributes(instanceID).UpdateOne(
		ctx,
		bson.M{"userId": userIDObj, "type": attributeType},
		bson.M{"$set": bson.M{"attributes": attributes, "createdAt": time.Now().UTC()}},
		options.Update().SetUpsert(true),
	)
	return err
}

// Delete all user attributes for a user
func (dbService *ParticipantUserDBService) DeleteAllUserAttributes(
	ctx context.Context,
	instanceID string,
	userID string,
) error {
	userIDObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = dbService.collectionParticipantUserAttributes(instanceID).DeleteMany(ctx, bson.M{"userId": userIDObj})
	return err
}

// Delete a user attribute for a user
func (dbService *ParticipantUserDBService) DeleteUserAttribute(instanceID string, userID string, attributeID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	userIDObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	attributeIDObj, err := primitive.ObjectIDFromHex(attributeID)
	if err != nil {
		return err
	}

	_, err = dbService.collectionParticipantUserAttributes(instanceID).DeleteOne(ctx, bson.M{"userId": userIDObj, "_id": attributeIDObj})
	return err
}

// Get all user attributes for a user
func (dbService *ParticipantUserDBService) GetAttributesForUser(instanceID string, userID string) ([]userTypes.UserAttributes, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	userIDObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := dbService.collectionParticipantUserAttributes(instanceID).Find(ctx, bson.M{"userId": userIDObj})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var userAttributes []userTypes.UserAttributes
	err = cursor.All(ctx, &userAttributes)
	if err != nil {
		return nil, err
	}
	return userAttributes, nil
}
