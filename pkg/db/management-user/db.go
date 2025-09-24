package managementuser

import (
	"context"
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collection names
const (
	COLLECTION_NAME_MANAGEMENT_USERS      = "management_users"
	COLLECTION_NAME_PERMISSIONS           = "permissions"
	COLLECTION_NAME_SESSIONS              = "management_user_sessions"
	COLLECTION_NAME_SERVICE_USERS         = "service_users"
	COLLECTION_NAME_SERVICE_USER_API_KEYS = "service_user_api_keys"
	COLLECTION_NAME_APP_ROLES             = "app_roles"
	COLLECTION_NAME_APP_ROLE_TEMPLATES    = "app_role_templates"
)

const (
	REMOVE_SESSIONS_AFTER = 60 * 60 * 24 * 2 // 2 days
)

type ManagementUserDBService struct {
	DBClient        *mongo.Client
	timeout         int
	noCursorTimeout bool
	DBNamePrefix    string
	InstanceIDs     []string
}

func NewManagementUserDBService(configs db.DBConfig) (*ManagementUserDBService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	defer cancel()

	dbClient, err := mongo.Connect(ctx,
		options.Client().ApplyURI(configs.URI),
		options.Client().SetMaxConnIdleTime(time.Duration(configs.IdleConnTimeout)*time.Second),
		options.Client().SetMaxPoolSize(configs.MaxPoolSize),
	)

	if err != nil {
		return nil, err
	}

	ctx, conCancel := context.WithTimeout(context.Background(), time.Duration(configs.Timeout)*time.Second)
	err = dbClient.Ping(ctx, nil)
	defer conCancel()

	if err != nil {
		return nil, err
	}

	muDBSc := &ManagementUserDBService{
		DBClient:        dbClient,
		timeout:         configs.Timeout,
		noCursorTimeout: configs.NoCursorTimeout,
		DBNamePrefix:    configs.DBNamePrefix,
		InstanceIDs:     configs.InstanceIDs,
	}

	return muDBSc, nil
}

func (dbService *ManagementUserDBService) getDBName(instanceID string) string {
	return dbService.DBNamePrefix + instanceID + "_users"
}

func (dbService *ManagementUserDBService) collectionManagementUsers(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_MANAGEMENT_USERS)
}

func (dbService *ManagementUserDBService) collectionPermissions(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_PERMISSIONS)
}

func (dbService *ManagementUserDBService) getContext() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(dbService.timeout)*time.Second)
}

func (dbService *ManagementUserDBService) CreateDefaultIndexes() {
	for _, instanceID := range dbService.InstanceIDs {
		start := time.Now()
		slog.Info("Creating default indexes for management user DB", slog.String("instanceID", instanceID))
		dbService.CreateDefaultIndexesForAppRolesCollection(instanceID)
		dbService.CreateDefaultIndexesForAppRoleTemplatesCollection(instanceID)
		dbService.CreateDefaultIndexesForManagementUsersCollection(instanceID)
		dbService.CreateDefaultIndexesForPermissionsCollection(instanceID)
		dbService.CreateDefaultIndexesForServiceUserAPIKeysCollection(instanceID)
		dbService.CreateDefaultIndexesForSessionsCollection(instanceID)
		slog.Info("Default indexes created for management user DB", slog.String("instanceID", instanceID), slog.String("duration", time.Since(start).String()))
	}
}

func (dbService *ManagementUserDBService) DropIndexes(dropAll bool) {
	for _, instanceID := range dbService.InstanceIDs {
		start := time.Now()
		slog.Info("Dropping indexes for management user DB", slog.String("instanceID", instanceID))
		dbService.DropIndexForAppRolesCollection(instanceID, dropAll)
		dbService.DropIndexForAppRoleTemplatesCollection(instanceID, dropAll)
		dbService.DropIndexForManagementUsersCollection(instanceID, dropAll)
		dbService.DropIndexForPermissionsCollection(instanceID, dropAll)
		dbService.DropIndexForServiceUserAPIKeysCollection(instanceID, dropAll)
		dbService.DropIndexForSessionsCollection(instanceID, dropAll)
		slog.Info("Indexes dropped for management user DB", slog.String("instanceID", instanceID), slog.String("duration", time.Since(start).String()))
	}
}
