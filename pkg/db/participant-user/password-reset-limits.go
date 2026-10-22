package participantuser

import (
	"errors"
	"log/slog"
	"sort"
	"time"

	userTypes "github.com/case-framework/case-backend/pkg/user-management/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PasswordResetCooldown    = 60      // seconds
	PasswordResetWindow      = 60 * 60 // rolling hour
	PasswordResetMaxRequests = 5
	passwordResetTTLIndex    = "expiresAt_1"
)

var ErrPasswordResetRateLimited = errors.New("password reset rate limited")

// One document per user in each instance's database. The built-in unique _id
// index serializes reservations and concurrent initialization for that user.
type passwordResetLimit struct {
	UserID     primitive.ObjectID `bson:"_id"`
	Timestamps []int64            `bson:"timestamps"`
	ExpiresAt  time.Time          `bson:"expiresAt"`
}

func (dbService *ParticipantUserDBService) CreatePasswordResetLimitIndexes(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()
	_, err := dbService.collectionPasswordResetLimits(instanceID).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetName(passwordResetTTLIndex).SetExpireAfterSeconds(0),
	})
	return err
}

func (dbService *ParticipantUserDBService) DropIndexForPasswordResetLimitsCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var err error
	if dropAll {
		_, err = dbService.collectionPasswordResetLimits(instanceID).Indexes().DropAll(ctx)
	} else {
		_, err = dbService.collectionPasswordResetLimits(instanceID).Indexes().DropOne(ctx, passwordResetTTLIndex)
	}
	if err != nil {
		slog.Error("Error dropping password reset indexes", slog.String("instanceID", instanceID), slog.String("error", err.Error()))
	}
}

// ReservePasswordReset loads the recipient and reserves quota before delivery.
// ErrNoDocuments means the account does not exist; a limit rejection has its own
// error. Delivery failures do not refund quota. No user document is modified.
func (dbService *ParticipantUserDBService) ReservePasswordReset(instanceID, accountID string) (userTypes.User, error) {
	return dbService.reservePasswordResetAt(instanceID, accountID, time.Now().Unix())
}

func (dbService *ParticipantUserDBService) reservePasswordResetAt(instanceID, accountID string, now int64) (userTypes.User, error) {
	user, err := dbService.GetUserByAccountID(instanceID, accountID)
	if err != nil {
		return userTypes.User{}, err
	}
	ctx, cancel := dbService.getContext()
	defer cancel()
	collection := dbService.collectionPasswordResetLimits(instanceID)

	// Seed only a missing document. Import recent legacy timestamps once so an
	// upgrade does not reset quota. Existing reservations always take precedence.
	history := make([]int64, 0, len(user.Account.PasswordResetTriggers))
	for _, ts := range user.Account.PasswordResetTriggers {
		if ts > now-PasswordResetWindow {
			history = append(history, ts)
		}
	}
	sort.Slice(history, func(i, j int) bool { return history[i] < history[j] })
	if len(history) > PasswordResetMaxRequests {
		history = history[len(history)-PasswordResetMaxRequests:]
	}
	expiresAt := now + PasswordResetWindow
	if len(history) > 0 {
		expiresAt = history[len(history)-1] + PasswordResetWindow
	}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$setOnInsert": passwordResetLimit{
		UserID: user.ID, Timestamps: history, ExpiresAt: time.Unix(expiresAt, 0),
	}}, options.Update().SetUpsert(true))
	// Concurrent first requests can race to create the same _id. The winner's
	// document is authoritative; losers still have to pass the atomic check below.
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return userTypes.User{}, err
	}

	recent := func(since int64) bson.M {
		return bson.M{"$filter": bson.M{
			"input": bson.M{"$ifNull": bson.A{"$timestamps", bson.A{}}},
			"as":    "ts", "cond": bson.M{"$gt": bson.A{"$$ts", since}},
		}}
	}
	filter := bson.M{"_id": user.ID, "$expr": bson.M{"$and": bson.A{
		bson.M{"$eq": bson.A{bson.M{"$size": recent(now - PasswordResetCooldown)}, 0}},
		bson.M{"$lt": bson.A{bson.M{"$size": recent(now - PasswordResetWindow)}, PasswordResetMaxRequests}},
	}}}
	update := mongo.Pipeline{bson.D{{Key: "$set", Value: bson.M{
		"timestamps": bson.M{"$concatArrays": bson.A{recent(now - PasswordResetWindow), bson.A{now}}},
		"expiresAt":  time.Unix(now+PasswordResetWindow, 0),
	}}}}
	// No upsert here: checking both limits and appending is a single atomic write.
	// A rejection does not change timestamps or extend TTL. If TTL deletes the
	// document between initialization and this write, fail closed for this request.
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return userTypes.User{}, err
	}
	if result.ModifiedCount != 1 {
		return userTypes.User{}, ErrPasswordResetRateLimited
	}
	return user, nil
}
