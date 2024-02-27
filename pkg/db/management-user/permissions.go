package managementuser

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create permission
func (dbService *ManagementUserDBService) CreatePermission(
	instanceID string,
	subjectID string,
	subjectType string,
	resourceType string,
	resourceKey string,
	action string,
	limiter string,
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
	cursor, err := dbService.collectionPermissions(instanceID).Find(ctx,
		bson.M{"subjectId": subjectID, "subjectType": subjectType, "resourceType": resourceType,
			"resourceKey": bson.M{"$in": resourceKey}, "action": action})
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

// Modify limiter of permission
func (dbService *ManagementUserDBService) UpdatePermissionLimiter(
	instanceID string,
	permissionID string,
	limiter string,
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
