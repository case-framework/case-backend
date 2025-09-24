package managementuser

import (
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (dbService *ManagementUserDBService) collectionSessions(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(COLLECTION_NAME_SESSIONS)
}

var indexesForSessionsCollection = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "createdAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(REMOVE_SESSIONS_AFTER).SetName("createdAt_1"),
	},
}

func (dbService *ManagementUserDBService) DropIndexForSessionsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionSessions(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for sessions", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForSessionsCollection {
			if index.Options.Name == nil {
				slog.Error("Index name is nil for sessions collection: ", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionSessions(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for sessions", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ManagementUserDBService) CreateDefaultIndexesForSessionsCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionSessions(instanceID).Indexes().CreateMany(ctx, indexesForSessionsCollection)
	if err != nil {
		slog.Error("Error creating index for sessions: ", slog.String("error", err.Error()))
	}
}

// Session represents a user session, created when a user logs in
func (dbService *ManagementUserDBService) CreateSession(
	instanceID string,
	userID string,
	renewToken string,
) (*Session, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	session := &Session{
		UserID:     userID,
		RenewToken: renewToken,
		CreatedAt:  time.Now(),
	}

	res, err := dbService.collectionSessions(instanceID).InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}
	session.ID = res.InsertedID.(primitive.ObjectID)
	return session, nil
}

// GetSession returns the session with the given ID
func (dbService *ManagementUserDBService) GetSession(
	instanceID string,
	sessionID string,
) (*Session, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var session Session
	objID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, err
	}
	err = dbService.collectionSessions(instanceID).FindOne(ctx, primitive.M{"_id": objID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteSession deletes the session with the given ID
func (dbService *ManagementUserDBService) DeleteSession(
	instanceID string,
	sessionID string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return err
	}
	_, err = dbService.collectionSessions(instanceID).DeleteOne(ctx, primitive.M{"_id": objID})
	return err
}

// DeleteSessionsByUserID deletes all sessions for the given user
func (dbService *ManagementUserDBService) DeleteSessionsByUserID(
	instanceID string,
	userID string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionSessions(instanceID).DeleteMany(ctx, primitive.M{"userId": userID})
	return err
}
