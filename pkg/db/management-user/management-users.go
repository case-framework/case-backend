package managementuser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (dbService *ManagementUserDBService) CreateUser(
	instanceID string,
	newUser *ManagementUser,
) (*ManagementUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	newUser.CreatedAt = time.Now()
	res, err := dbService.collectionManagementUsers(instanceID).InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}
	newUser.ID = res.InsertedID.(primitive.ObjectID)
	return newUser, nil
}

// find user by sub
func (dbService *ManagementUserDBService) GetUserBySub(
	instanceID string,
	sub string,
) (*ManagementUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	var user ManagementUser
	err := dbService.collectionManagementUsers(instanceID).FindOne(ctx, bson.M{"sub": sub}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// find user by id
func (dbService *ManagementUserDBService) GetUserByID(
	instanceID string,
	id string,
) (*ManagementUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	var user ManagementUser
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = dbService.collectionManagementUsers(instanceID).FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// update user
func (dbService *ManagementUserDBService) UpdateUser(
	instanceID string,
	id string,
	email string,
	username string,
	isAdmin bool,
	lastLogin time.Time,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = dbService.collectionManagementUsers(instanceID).UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"email":       email,
				"username":    username,
				"isAdmin":     isAdmin,
				"lastLoginAt": lastLogin,
			},
		},
	)
	return err
}

// delete user
func (dbService *ManagementUserDBService) DeleteUser(
	instanceID string,
	id string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = dbService.collectionManagementUsers(instanceID).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
