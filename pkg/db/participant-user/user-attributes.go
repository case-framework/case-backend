package participantuser

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

const idxParticipantUserAttributesUserIdType = "userId_1_type_1"

var defaultParticipantUserAttributeIndexNames = []string{
	idxParticipantUserAttributesUserIdType,
}

var indexesForParticipantUserAttributesCollection = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "type", Value: 1}},
		Options: options.Index().SetName(idxParticipantUserAttributesUserIdType).SetUnique(true),
	},
}

func (dbService *ParticipantUserDBService) DropIndexForParticipantUserAttributesCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for participant user attributes", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, indexName := range defaultParticipantUserAttributeIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for participant user attributes collection", slog.String("instanceID", instanceID))
				continue
			}
			err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for participant user attributes", slog.String("error", err.Error()), slog.String("indexName", indexName), slog.String("instanceID", instanceID))
			}
		}
	}
}

func (dbService *ParticipantUserDBService) CreateDefaultIndexesForParticipantUserAttributesCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().CreateMany(ctx, indexesForParticipantUserAttributesCollection)
	if err != nil {
		slog.Error("Error creating index for participant user attributes", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
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

	userIDObj, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = dbService.collectionParticipantUserAttributes(instanceID).UpdateOne(
		ctx,
		bson.M{"userId": userIDObj, "type": attributeType},
		bson.M{"$set": bson.M{"attributes": attributes, "createdAt": time.Now().UTC()}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

// Delete all user attributes for a user
func (dbService *ParticipantUserDBService) DeleteAllUserAttributes(
	ctx context.Context,
	instanceID string,
	userID string,
) error {
	userIDObj, err := bson.ObjectIDFromHex(userID)
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

	userIDObj, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	attributeIDObj, err := bson.ObjectIDFromHex(attributeID)
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

	userIDObj, err := bson.ObjectIDFromHex(userID)
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
