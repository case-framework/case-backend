package participantuser

import (
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	FAILED_OTP_ATTEMP_WINDOW      = 60 * 5
	idxFailedOtpAttemptsUserID    = "userID_1"
	idxFailedOtpAttemptsTimestamp = "timestamp_1"
)

var defaultFailedOtpAttemptIndexNames = []string{
	idxFailedOtpAttemptsUserID,
	idxFailedOtpAttemptsTimestamp,
}

type FailedOtpAttempt struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	UserID    string    `json:"userId" bson:"userID"`
}

var indexesForFailedOtpAttemptsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userID", Value: 1},
		},
		Options: options.Index().SetName(idxFailedOtpAttemptsUserID),
	},
	{
		Keys: bson.D{
			{Key: "timestamp", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(FAILED_OTP_ATTEMP_WINDOW).SetName(idxFailedOtpAttemptsTimestamp),
	},
}

func (dbService *ParticipantUserDBService) DropIndexForFailedOtpAttemptsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		err := dbService.collectionFailedOtpAttempts(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for FailedOtpAttempts", slog.String("error", err.Error()))
		}
	} else {
		for _, indexName := range defaultFailedOtpAttemptIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for FailedOtpAttempts collection")
				continue
			}
			err := dbService.collectionFailedOtpAttempts(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for FailedOtpAttempts", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ParticipantUserDBService) CreateDefaultIndexesForFailedOtpAttemptsCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionFailedOtpAttempts(instanceID).Indexes().CreateMany(ctx, indexesForFailedOtpAttemptsCollection)
	if err != nil {
		slog.Error("Error creating index for FailedOtpAttempts", slog.String("error", err.Error()))
	}
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
