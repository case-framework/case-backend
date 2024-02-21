package managementuser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (dbService *ManagementUserDBService) collectionSessions(instanceID string) *mongo.Collection {
	return dbService.DBClient.Database(dbService.getDBName(instanceID)).Collection(Sessions)
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
