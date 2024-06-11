package participantuser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	FAILED_OTP_ATTEMP_WINDOW = 60 * 5
)

type FailedOtpAttempt struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	UserID    string    `json:"userId" bson:"userID"`
}

func (dbService *ParticipantUserDBService) CreateIndexForFailedOtpAttempts(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionFailedOtpAttempts(instanceID).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "userID", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "timestamp", Value: 1},
				},
				Options: options.Index().SetExpireAfterSeconds(FAILED_OTP_ATTEMP_WINDOW),
			},
		},
	)
	return err
}

func (dbService *ParticipantUserDBService) CountFailedOtpAttempts(instanceID string, userID string) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userID": userID,
		"timestamp": bson.M{
			"$gt": time.Now().Add(-FAILED_OTP_ATTEMP_WINDOW * time.Second),
		},
	}
	return dbService.collectionFailedOtpAttempts(instanceID).CountDocuments(ctx, filter)
}

func (dbService *ParticipantUserDBService) AddFailedOtpAttempt(instanceID string, userID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionFailedOtpAttempts(instanceID).InsertOne(ctx, FailedOtpAttempt{
		Timestamp: time.Now(),
		UserID:    userID,
	})
	return err
}
