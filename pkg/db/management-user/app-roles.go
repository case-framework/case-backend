package managementuser

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// store which users have which roles
func (dbService *ManagementUserDBService) collectionAppRoles(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_APP_ROLES)
}

// store the templates for app roles, to be able to add new roles from templates
func (dbService *ManagementUserDBService) collectionAppRoleTemplates(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_APP_ROLE_TEMPLATES)
}

func (dbService *ManagementUserDBService) createIndexForAppRoles(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if _, err := dbService.collectionAppRoles(instanceID).Indexes().DropAll(ctx); err != nil {
		slog.Error("Error dropping indexes for permissions: ", slog.String("error", err.Error()))
	}

	_, err := dbService.collectionAppRoles(instanceID).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "subjectID", Value: 1},
			},
		},
	)
	return err
}

/// App role templates

// Add a new app role template

// Get all app role templates

// Get a app role template by id

// Update a app role template

// Delete a app role template

/// App roles

// Add a new app role for a user
func (dbService *ManagementUserDBService) AddAppRoleForSubject(
	instanceID string,
	subjectID string,
	subjectType string,
	appName string,
	role string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	appRole := &AppRole{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		AppName:     appName,
		Role:        role,
		CreatedAt:   time.Now(),
	}

	res, err := dbService.collectionAppRoles(instanceID).InsertOne(ctx, appRole)
	if err != nil {
		return err
	}
	appRole.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// Get all app roles
func (dbService *ManagementUserDBService) GetAllAppRoles(
	instanceID string,
) ([]AppRole, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var appRoles []AppRole
	cursor, err := dbService.collectionAppRoles(instanceID).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &appRoles); err != nil {
		return nil, err
	}

	return appRoles, nil
}

// Get app roles for a user
func (dbService *ManagementUserDBService) GetAppRolesForSubject(
	instanceID string,
	subjectID string,
) ([]AppRole, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	cursor, err := dbService.collectionAppRoles(instanceID).Find(ctx, bson.M{"subjectId": subjectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appRoles []AppRole
	if err := cursor.All(ctx, &appRoles); err != nil {
		return nil, err
	}

	return appRoles, nil
}

// Remove an app role
func (dbService *ManagementUserDBService) DeleteAppRole(
	instanceID string,
	appRoleID string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(appRoleID)
	if err != nil {
		return err
	}
	_, err = dbService.collectionAppRoles(instanceID).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// Remove all app roles for a user
func (dbService *ManagementUserDBService) RemoveAllAppRolesForSubject(
	instanceID string,
	subjectID string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionAppRoles(instanceID).DeleteMany(ctx, bson.M{"subjectId": subjectID})
	return err
}

// Remove all app roles for an app
func (dbService *ManagementUserDBService) RemoveAllAppRolesForApp(
	instanceID string,
	appName string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionAppRoles(instanceID).DeleteMany(ctx, bson.M{"appName": appName})
	return err
}
