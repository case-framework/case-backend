package participantuser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	umTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

func (dbService *ParticipantUserDBService) CreateIndexForParticipantUsers(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionParticipantUsers(instanceID).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "timestamps.markedForDeletion", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "account.accountID", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "timestamps.createdAt", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "account.accountConfirmedAt", Value: 1},
					{Key: "timestamps.createdAt", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "contactPreferences.receiveWeeklyMessageDayOfWeek", Value: 1},
				},
			},
		},
	)
	return err
}

func (dbService *ParticipantUserDBService) GetUserByAccountID(instanceID, accountID string) (umTypes.User, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var user umTypes.User
	filter := bson.M{"account.accountID": accountID}
	err := dbService.collectionParticipantUsers(instanceID).FindOne(ctx, filter).Decode(&user)
	return user, err
}

func (dbService *ParticipantUserDBService) SaveFailedLoginAttempt(instanceID string, userID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": _id}
	update := bson.M{
		"$push": bson.M{
			"account.failedLoginAttempts": time.Now().Unix(),
		},
	}
	_, err = dbService.collectionParticipantUsers(instanceID).UpdateOne(ctx, filter, update)
	return err
}

// low level find and replace
func (dbService *ParticipantUserDBService) _updateUserInDB(orgID string, user umTypes.User) (umTypes.User, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	elem := umTypes.User{}
	filter := bson.M{"_id": user.ID}
	rd := options.After
	fro := options.FindOneAndReplaceOptions{
		ReturnDocument: &rd,
	}
	err := dbService.collectionParticipantUsers(orgID).FindOneAndReplace(ctx, filter, user, &fro).Decode(&elem)
	return elem, err
}

func (dbService *ParticipantUserDBService) UpdateUser(instanceID string, updatedUser umTypes.User) (umTypes.User, error) {
	// Set last update time
	updatedUser.Timestamps.UpdatedAt = time.Now().Unix()
	return dbService._updateUserInDB(instanceID, updatedUser)
}
