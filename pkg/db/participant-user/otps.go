package participantuser

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
)

const (
	OTP_TTL = 60 * 15
)

func (dbService *ParticipantUserDBService) CreateIndexForOTPs(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionOTPs(instanceID).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "userID", Value: 1},
					{Key: "code", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: "createdAt", Value: 1},
				},
				Options: options.Index().SetExpireAfterSeconds(OTP_TTL),
			},
		},
	)
	return err
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
