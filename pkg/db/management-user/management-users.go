package managementuser

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var managementUserIndexNames []string

var indexesForManagementUsersCollection = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "sub", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_sub_1"),
	},
}

func (dbService *ManagementUserDBService) DropIndexForManagementUsersCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		err := dbService.collectionManagementUsers(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for management users", slog.String("error", err.Error()))
		}
	} else {
		for _, indexName := range managementUserIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for management users collection", slog.String("instanceID", instanceID))
				continue
			}
			err := dbService.collectionManagementUsers(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for management users", slog.String("error", err.Error()), slog.String("indexName", indexName), slog.String("instanceID", instanceID))
			}
		}
	}
}

func (dbService *ManagementUserDBService) CreateDefaultIndexesForManagementUsersCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	names, err := dbService.collectionManagementUsers(instanceID).Indexes().CreateMany(ctx, indexesForManagementUsersCollection)
	if err != nil {
		slog.Error("Error creating index for management users", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
	managementUserIndexNames = names
}

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
	newUser.ID = res.InsertedID.(bson.ObjectID)
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
	objID, err := bson.ObjectIDFromHex(id)
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
	provider string,
	isAdmin bool,
	lastLogin time.Time,
	imageURL string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	objID, err := bson.ObjectIDFromHex(id)
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
				"provider":    provider,
				"isAdmin":     isAdmin,
				"lastLoginAt": lastLogin,
				"imageUrl":    imageURL,
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

	// delete all app roles for the user
	err := dbService.RemoveAllAppRolesForSubject(instanceID, id)
	if err != nil {
		return err
	}

	// delete the user
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = dbService.collectionManagementUsers(instanceID).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// get all management users
func (dbService *ManagementUserDBService) GetAllUsers(
	instanceID string,
	returnFullObject bool,
) ([]*ManagementUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.D{}

	opts := options.Find()
	if !returnFullObject {
		opts = options.Find().SetProjection(bson.D{
			{Key: "_id", Value: 1},
			{Key: "email", Value: 1},
			{Key: "username", Value: 1},
			{Key: "provider", Value: 1},
			{Key: "isAdmin", Value: 1},
			{Key: "imageUrl", Value: 1},
		})
	}

	var users []*ManagementUser
	cursor, err := dbService.collectionManagementUsers(instanceID).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var user ManagementUser
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

// Get users by ids
func (dbService *ManagementUserDBService) GetUsersByIDs(
	instanceID string,
	ids []string,
	returnFullObject bool,
) ([]*ManagementUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var objIDs []bson.ObjectID
	for _, id := range ids {
		objID, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objIDs = append(objIDs, objID)
	}

	filter := bson.M{"_id": bson.M{"$in": objIDs}}

	opts := options.Find()
	if !returnFullObject {
		opts = options.Find().SetProjection(bson.D{
			{Key: "_id", Value: 1},
			{Key: "email", Value: 1},
			{Key: "username", Value: 1},
			{Key: "provider", Value: 1},
			{Key: "isAdmin", Value: 1},
			{Key: "imageUrl", Value: 1},
		})
	}

	var users []*ManagementUser
	cursor, err := dbService.collectionManagementUsers(instanceID).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var user ManagementUser
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
