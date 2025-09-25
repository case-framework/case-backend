package managementuser

import (
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var indexesForPermissionsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "subjectId", Value: 1},
			{Key: "subjectType", Value: 1},
			{Key: "resourceType", Value: 1},
			{Key: "resourceKey", Value: 1},
			{Key: "action", Value: 1},
		},
		Options: options.Index().SetName("subjectId_1_subjectType_1_resourceType_1_resourceKey_1_action_1"),
	},
}

func (dbService *ManagementUserDBService) DropIndexForPermissionsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionPermissions(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for permissions: ", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForPermissionsCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for permissions collection: ", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionPermissions(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for permissions: ", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ManagementUserDBService) CreateDefaultIndexesForPermissionsCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionPermissions(instanceID).Indexes().CreateMany(ctx, indexesForPermissionsCollection)
	if err != nil {
		slog.Error("Error creating index for permissions: ", slog.String("error", err.Error()))
	}
}

// Create permission
func (dbService *ManagementUserDBService) CreatePermission(
	instanceID string,
	subjectID string,
	subjectType string,
	resourceType string,
	resourceKey string,
	action string,
	limiter []map[string]string,
) (*Permission, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	permission := &Permission{
		SubjectID:    subjectID,
		SubjectType:  subjectType,
		ResourceType: resourceType,
		ResourceKey:  resourceKey,
		Action:       action,
		Limiter:      limiter,
	}

	res, err := dbService.collectionPermissions(instanceID).InsertOne(ctx, permission)
	if err != nil {
		return nil, err
	}
	permission.ID = res.InsertedID.(primitive.ObjectID)
	return permission, nil
}

// Find permission by id
func (dbService *ManagementUserDBService) GetPermissionByID(
	instanceID string,
	permissionID string,
) (*Permission, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		return nil, err
	}
	var permission Permission
	if err := dbService.collectionPermissions(instanceID).FindOne(ctx, bson.M{"_id": objID}).Decode(&permission); err != nil {
		return nil, err
	}
	return &permission, nil
}

// Find permissions by subject id and type
func (dbService *ManagementUserDBService) GetPermissionBySubject(
	instanceID string,
	subjectID string,
	subjectType string,
) ([]*Permission, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var permissions []*Permission
	cursor, err := dbService.collectionPermissions(instanceID).Find(ctx, bson.M{"subjectId": subjectID, "subjectType": subjectType})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var permission Permission
		if err := cursor.Decode(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}
	return permissions, nil
}

// Find permissions by subject id and type and resource type
func (dbService *ManagementUserDBService) GetPermissionBySubjectAndResourceForAction(
	instanceID string,
	subjectID string,
	subjectType string,
	resourceType string,
	resourceKey []string,
	action string,
) ([]*Permission, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var permissions []*Permission

	actions := []string{action}
	if action != "*" {
		actions = append(actions, "*")
	}
	cursor, err := dbService.collectionPermissions(instanceID).Find(ctx,
		bson.M{"subjectId": subjectID, "subjectType": subjectType, "resourceType": resourceType,
			"resourceKey": bson.M{"$in": resourceKey}, "action": bson.M{"$in": actions}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var permission Permission
		if err := cursor.Decode(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}
	return permissions, nil
}

// Find permissions by resource type and key
func (dbService *ManagementUserDBService) GetPermissionByResource(
	instanceID string,
	resourceType string,
	resourceKey string,
) ([]*Permission, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var permissions []*Permission
	cursor, err := dbService.collectionPermissions(instanceID).Find(ctx, bson.M{"resourceType": resourceType, "resourceKey": resourceKey})
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}

// Modify limiter of permission
func (dbService *ManagementUserDBService) UpdatePermissionLimiter(
	instanceID string,
	permissionID string,
	limiter []map[string]string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		return err
	}
	_, err = dbService.collectionPermissions(instanceID).UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{"limiter": limiter},
		},
	)
	return err
}

// Delete permission
func (dbService *ManagementUserDBService) DeletePermission(
	instanceID string,
	permissionID string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(permissionID)
	if err != nil {
		return err
	}
	_, err = dbService.collectionPermissions(instanceID).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// Delete permissions by subject id and type
func (dbService *ManagementUserDBService) DeletePermissionsBySubject(
	instanceID string,
	subjectID string,
	subjectType string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionPermissions(instanceID).DeleteMany(ctx, bson.M{"subjectId": subjectID, "subjectType": subjectType})
	return err
}
