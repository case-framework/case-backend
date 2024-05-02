package managementuser

import (
	"context"
	"log/slog"
	"time"

	"github.com/case-framework/case-backend/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collection names
const (
	COLLECTION_NAME_MANAGEMENT_USERS = "management_users"
	COLLECTION_NAME_PERMISSIONS      = "permissions"
	COLLECTION_NAME_SESSIONS         = "management_user_sessions"
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

	if configs.RunIndexCreation {
		if err := muDBSc.ensureIndexes(); err != nil {
			slog.Error("Error ensuring indexes for management user DB: ", err)
		}
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

func (dbService *ManagementUserDBService) ensureIndexes() error {
	slog.Debug("Ensuring indexes for management user DB")
	for _, instanceID := range dbService.InstanceIDs {
		ctx, cancel := dbService.getContext()
		defer cancel()

		// create unique index for sub
		_, err := dbService.collectionManagementUsers(instanceID).Indexes().CreateOne(
			ctx,
			mongo.IndexModel{
				Keys:    bson.D{{Key: "sub", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		)
		if err != nil {
			slog.Error("Error creating unique index for sub in userDB.management_users: ", err)
		}

		// create index for permissions
		_, err = dbService.collectionPermissions(instanceID).Indexes().CreateOne(
			ctx,
			mongo.IndexModel{
				Keys: bson.D{
					{Key: "subjectID", Value: 1},
					{Key: "subjectType", Value: 1},
					{Key: "resourceType", Value: 1},
					{Key: "resourceID", Value: 1},
					{Key: "action", Value: 1},
				},
			},
		)
		if err != nil {
			slog.Error("Error creating index for permissions in userDB.permissions: ", err)
		}

		// create index for sessions
		_, err = dbService.collectionSessions(instanceID).Indexes().CreateOne(
			ctx,
			mongo.IndexModel{
				Keys:    bson.D{{Key: "createdAt", Value: 1}},
				Options: options.Index().SetExpireAfterSeconds(REMOVE_SESSIONS_AFTER),
			},
		)
		if err != nil {
			slog.Error("Error creating index for createdAt in userDB.sessions: ", err)
		}
	}

	return nil
}
