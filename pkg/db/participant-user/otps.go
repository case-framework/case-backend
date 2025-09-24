package participantuser

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

const (
	OTP_TTL = 60 * 15
)

var indexesForOTPsCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "userID", Value: 1},
			{Key: "code", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("userID_code_1"),
	},
	{
		Keys: bson.D{
			{Key: "createdAt", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(OTP_TTL).SetName("createdAt_1"),
	},
}

func (dbService *ParticipantUserDBService) DropIndexForOTPsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionOTPs(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for OTPs", slog.String("error", err.Error()))
		}
	} else {
		for _, index := range indexesForOTPsCollection {
			if index.Options.Name == nil {
				slog.Error("Index name is nil for OTPs collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionOTPs(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for OTPs", slog.String("error", err.Error()), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *ParticipantUserDBService) CreateDefaultIndexesForOTPsCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionOTPs(instanceID).Indexes().CreateMany(ctx, indexesForOTPsCollection)
	if err != nil {
		slog.Error("Error creating index for OTPs", slog.String("error", err.Error()))
	}
}

func (dbService *ParticipantUserDBService) CreateOTP(instanceID string, userID string, code string, t userTypes.OTPType, maxOTPCount int64) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	session, err := dbService.collectionOTPs(instanceID).Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	createOTPIfLimitNotReached := func(sessCtx mongo.SessionContext) error {

		filter := bson.M{"userID": userID}
		count, err := dbService.collectionOTPs(instanceID).CountDocuments(sessCtx, filter)
		if err != nil {
			return err
		}

		if count >= maxOTPCount {
			return errors.New("too many OTP requests")
		}

		otp := userTypes.OTP{
			UserID:    userID,
			Code:      code,
			Type:      t,
			CreatedAt: time.Now(),
		}
		_, err = dbService.collectionOTPs(instanceID).InsertOne(sessCtx, otp)
		return err
	}

	return mongo.WithSession(ctx, session, createOTPIfLimitNotReached)
}

func (dbService *ParticipantUserDBService) FindOTP(instanceID string, userID string, code string) (userTypes.OTP, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userID": userID, "code": code}
	var otp userTypes.OTP
	err := dbService.collectionOTPs(instanceID).FindOne(ctx, filter).Decode(&otp)
	return otp, err
}

func (dbService *ParticipantUserDBService) DeleteOTP(instanceID string, userID string, code string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userID": userID, "code": code}
	_, err := dbService.collectionOTPs(instanceID).DeleteOne(ctx, filter)
	return err
}

func (dbService *ParticipantUserDBService) DeleteOTPs(instanceID string, userID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userID": userID}
	_, err := dbService.collectionOTPs(instanceID).DeleteMany(ctx, filter)
	return err
}

func (dbService *ParticipantUserDBService) CountOTP(instanceID string, userID string) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userID": userID}
	count, err := dbService.collectionOTPs(instanceID).CountDocuments(ctx, filter)
	return count, err
}

func (dbService *ParticipantUserDBService) GetLastOTP(instanceID string, userID string, otpType string) (userTypes.OTP, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userID": userID, "type": otpType}
	sort := bson.M{"createdAt": -1}

	var otp userTypes.OTP
	err := dbService.collectionOTPs(instanceID).FindOne(ctx, filter, options.FindOne().SetSort(sort)).Decode(&otp)
	return otp, err
}
