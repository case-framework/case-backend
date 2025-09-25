package managementuser

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *ManagementUserDBService) collectionServiceUsers(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_SERVICE_USERS)
}

func (dbService *ManagementUserDBService) collectionServiceUserAPIKeys(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_SERVICE_USER_API_KEYS)
}

var indexesForServiceUserAPIKeysCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "key", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("key_1"),
	},
	{
		Keys: bson.D{
			{Key: "expiresAt", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(0).SetName("expiresAt_1"),
	},
}

func (dbService *ManagementUserDBService) DropIndexForServiceUserAPIKeysCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionServiceUserAPIKeys(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for service user API keys: ", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForServiceUserAPIKeysCollection {
			if index.Options.Name == nil {
				slog.Error("Index name is nil for service user API keys collection: ", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionServiceUserAPIKeys(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for service user API keys: ", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ManagementUserDBService) CreateDefaultIndexesForServiceUserAPIKeysCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionServiceUserAPIKeys(instanceID).Indexes().CreateMany(ctx, indexesForServiceUserAPIKeysCollection)
	if err != nil {
		slog.Error("Error creating index for service user API keys: ", slog.String("error", err.Error()))
	}
}

// CreateServiceUser creates a new service user
func (dbService *ManagementUserDBService) CreateServiceUser(instanceID string, label string, description string) (*ServiceUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	serviceUser := &ServiceUser{
		Label:       label,
		Description: description,
		CreatedAt:   time.Now(),
	}

	result, err := dbService.collectionServiceUsers(instanceID).InsertOne(ctx, serviceUser)
	if err != nil {
		slog.Error("Error creating service user", slog.String("error", err.Error()))
		return nil, err
	}

	if result.InsertedID == nil {
		err = errors.New("InsertedID is nil")
		slog.Error("Error creating service user", slog.String("error", err.Error()))
		return nil, err
	}

	serviceUser.ID = result.InsertedID.(primitive.ObjectID)

	return serviceUser, nil
}

// GetServiceUserByID returns a service user by its ID
func (dbService *ManagementUserDBService) GetServiceUserByID(instanceID string, id string) (*ServiceUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var serviceUser ServiceUser
	err = dbService.collectionServiceUsers(instanceID).FindOne(ctx, bson.M{"_id": _id}).Decode(&serviceUser)
	if err != nil {
		slog.Error("Error getting service user by ID", slog.String("error", err.Error()))
		return nil, err
	}

	return &serviceUser, nil
}

// GetServiceUsers returns all service users
func (dbService *ManagementUserDBService) GetServiceUsers(instanceID string) ([]ServiceUser, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var serviceUsers []ServiceUser
	cursor, err := dbService.collectionServiceUsers(instanceID).Find(ctx, bson.M{})
	if err != nil {
		slog.Error("Error getting service users", slog.String("error", err.Error()))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var serviceUser ServiceUser
		err := cursor.Decode(&serviceUser)
		if err != nil {
			slog.Error("Error decoding service user", slog.String("error", err.Error()))
			return nil, err
		}
		serviceUsers = append(serviceUsers, serviceUser)
	}

	return serviceUsers, nil
}

// DeleteServiceUser deletes a service user by its ID and all its API keys
func (dbService *ManagementUserDBService) DeleteServiceUser(instanceID string, id string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("Error converting ID to ObjectID", slog.String("error", err.Error()))
		return err
	}

	_, err = dbService.collectionServiceUserAPIKeys(instanceID).DeleteMany(ctx, bson.M{"serviceUserId": id})
	if err != nil {
		slog.Error("Error deleting service user API keys", slog.String("error", err.Error()))
		return err
	}

	err = dbService.DeletePermissionsBySubject(instanceID, id, "service-account")
	if err != nil {
		slog.Error("Error deleting service user permissions", slog.String("error", err.Error()))
		return err
	}

	_, err = dbService.collectionServiceUsers(instanceID).DeleteOne(ctx, bson.M{"_id": _id})
	if err != nil {
		slog.Error("Error deleting service user", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// UpdateServiceUser updates a service user by its ID
func (dbService *ManagementUserDBService) UpdateServiceUser(instanceID string, id string, label string, description string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("Error converting ID to ObjectID", slog.String("error", err.Error()))
		return err
	}

	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"label": label, "description": description}}

	_, err = dbService.collectionServiceUsers(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		slog.Error("Error updating service user", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (dbService *ManagementUserDBService) CreateServiceUserAPIKey(instanceID string, serviceUserID string, apiKey string, expiresAt *time.Time) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	sApiKey := ServiceUserAPIKey{
		ServiceUserID: serviceUserID,
		Key:           apiKey,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now(),
	}

	_, err := dbService.collectionServiceUserAPIKeys(instanceID).InsertOne(ctx, sApiKey)
	if err != nil {
		slog.Error("Error creating service user API key", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (dbService *ManagementUserDBService) UpdateServiceUserAPIKeyLastUsedAt(instanceID string, apiKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"key": apiKey}
	update := bson.M{"$set": bson.M{"lastUsedAt": time.Now()}}

	_, err := dbService.collectionServiceUserAPIKeys(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		slog.Error("Error updating service user API key last used at", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (dbService *ManagementUserDBService) GetServiceUserAPIKey(instanceID string, apiKey string) (*ServiceUserAPIKey, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var sApiKey ServiceUserAPIKey

	err := dbService.collectionServiceUserAPIKeys(instanceID).FindOne(ctx, bson.M{"key": apiKey}).Decode(&sApiKey)
	if err != nil {
		slog.Error("Error getting service user API key", slog.String("error", err.Error()))
		return nil, err
	}

	err = dbService.UpdateServiceUserAPIKeyLastUsedAt(instanceID, apiKey)
	if err != nil {
		slog.Error("Error updating service user API key last used at", slog.String("error", err.Error()))
		return nil, err
	}

	return &sApiKey, nil
}

func (dbService *ManagementUserDBService) DeleteServiceUserAPIKey(instanceID string, id string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": _id}
	_, err = dbService.collectionServiceUserAPIKeys(instanceID).DeleteOne(ctx, filter)
	if err != nil {
		slog.Error("Error deleting service user API key", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (dbService *ManagementUserDBService) GetServiceUserAPIKeys(instanceID string, serviceUserID string) ([]ServiceUserAPIKey, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var sApiKeys []ServiceUserAPIKey

	filter := bson.M{"serviceUserId": serviceUserID}
	cursor, err := dbService.collectionServiceUserAPIKeys(instanceID).Find(ctx, filter)
	if err != nil {
		slog.Error("Error getting service user API keys", slog.String("error", err.Error()))
		return nil, err
	}

	err = cursor.All(ctx, &sApiKeys)
	if err != nil {
		slog.Error("Error getting service user API keys", slog.String("error", err.Error()))
		return nil, err
	}

	return sApiKeys, nil
}
