package participantuser

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

func (dbService *ParticipantUserDBService) CreateIndexForParticipantUserAttributes(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	const idxName = "idx_user_attributes_userId"

	if _, err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().DropOne(ctx, idxName); err != nil {
		// Index might not exist yet; log at debug level
		slog.Debug("Drop index for participant user attributes", slog.String("index", idxName), slog.String("error", err.Error()))
	}

	_, err := dbService.collectionParticipantUserAttributes(instanceID).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().SetName(idxName),
		},
	)
	return err
}

// Create a user attribute for a user
func (dbService *ParticipantUserDBService) CreateUserAttribute(
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

	_, err = dbService.collectionParticipantUserAttributes(instanceID).InsertOne(ctx, userTypes.UserAttributes{
		UserID:     userIDObj,
		Type:       attributeType,
		Attributes: attributes,
		CreatedAt:  time.Now().UTC(),
	})
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
