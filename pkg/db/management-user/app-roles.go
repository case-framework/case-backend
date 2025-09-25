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

// store which users have which roles
func (dbService *ManagementUserDBService) collectionAppRoles(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_APP_ROLES)
}

// store the templates for app roles, to be able to add new roles from templates
func (dbService *ManagementUserDBService) collectionAppRoleTemplates(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_APP_ROLE_TEMPLATES)
}

var indexesForAppRolesCollection = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "subjectId", Value: 1}},
		Options: options.Index().SetName("subjectId_1"),
	},
	{
		Keys:    bson.D{{Key: "appName", Value: 1}},
		Options: options.Index().SetName("appName_1"),
	},
	{
		Keys: bson.D{
			{Key: "subjectType", Value: 1},
			{Key: "subjectId", Value: 1},
			{Key: "appName", Value: 1},
			{Key: "role", Value: 1},
		},
		Options: options.Index().SetName("uniq_subjectType_1_subjectId_1_appName_1_role_1").SetUnique(true),
	},
}

func (dbService *ManagementUserDBService) DropIndexForAppRolesCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	if dropAll {
		_, err := dbService.collectionAppRoles(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for app roles", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForAppRolesCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for app roles collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionAppRoles(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for app roles", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ManagementUserDBService) CreateDefaultIndexesForAppRolesCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionAppRoles(instanceID).Indexes().CreateMany(ctx, indexesForAppRolesCollection)
	if err != nil {
		slog.Error("Error creating index for app roles", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

var indexesForAppRoleTemplatesCollection = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "appName", Value: 1}, {Key: "role", Value: 1}},
		Options: options.Index().SetName("uniq_appName_1_role_1").SetUnique(true),
	},
	{
		Keys:    bson.D{{Key: "appName", Value: 1}},
		Options: options.Index().SetName("appName_1"),
	},
}

func (dbService *ManagementUserDBService) DropIndexForAppRoleTemplatesCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	if dropAll {
		_, err := dbService.collectionAppRoleTemplates(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for app role templates", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForAppRoleTemplatesCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for app role templates collection: ", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionAppRoleTemplates(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for app role templates", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ManagementUserDBService) CreateDefaultIndexesForAppRoleTemplatesCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionAppRoleTemplates(instanceID).Indexes().CreateMany(ctx, indexesForAppRoleTemplatesCollection)
	if err != nil {
		slog.Error("Error creating index for app role templates", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

/// App role templates

// Add a new app role template
func (dbService *ManagementUserDBService) AddAppRoleTemplate(
	instanceID string,
	appName string,
	role string,
	requiredPermissions []Permission,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	appRoleTemplate := AppRoleTemplate{
		AppName:             appName,
		Role:                role,
		RequiredPermissions: requiredPermissions,
		CreatedAt:           time.Now(),
	}
	res, err := dbService.collectionAppRoleTemplates(instanceID).InsertOne(ctx, appRoleTemplate)
	if err != nil {
		return err
	}
	appRoleTemplate.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// Get all app role templates
func (dbService *ManagementUserDBService) GetAllAppRoleTemplates(
	instanceID string,
) ([]AppRoleTemplate, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var appRoleTemplates []AppRoleTemplate
	cursor, err := dbService.collectionAppRoleTemplates(instanceID).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &appRoleTemplates); err != nil {
		return nil, err
	}

	return appRoleTemplates, nil
}

// Get a app role template by id
func (dbService *ManagementUserDBService) GetAppRoleTemplateByID(
	instanceID string,
	appRoleTemplateID string,
) (AppRoleTemplate, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(appRoleTemplateID)
	if err != nil {
		return AppRoleTemplate{}, err
	}

	var appRoleTemplate AppRoleTemplate
	if err := dbService.collectionAppRoleTemplates(instanceID).FindOne(ctx, bson.M{"_id": objID}).Decode(&appRoleTemplate); err != nil {
		return AppRoleTemplate{}, err
	}

	return appRoleTemplate, nil
}

// Update a app role template
func (dbService *ManagementUserDBService) UpdateAppRoleTemplate(
	instanceID string,
	appRoleTemplateID string,
	appName string,
	role string,
	requiredPermissions []Permission,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(appRoleTemplateID)
	if err != nil {
		return err
	}

	res, err := dbService.collectionAppRoleTemplates(instanceID).UpdateOne(ctx, bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"appName":             appName,
			"role":                role,
			"requiredPermissions": requiredPermissions,
			"updatedAt":           time.Now(),
		}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("app role template not found")
	}
	return nil
}

// Delete a app role template
func (dbService *ManagementUserDBService) DeleteAppRoleTemplate(
	instanceID string,
	appRoleTemplateID string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(appRoleTemplateID)
	if err != nil {
		return err
	}
	_, err = dbService.collectionAppRoleTemplates(instanceID).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// Remove all app role templates for an app
func (dbService *ManagementUserDBService) RemoveAllAppRoleTemplatesForApp(
	instanceID string,
	appName string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionAppRoleTemplates(instanceID).DeleteMany(ctx, bson.M{"appName": appName})
	return err
}

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
