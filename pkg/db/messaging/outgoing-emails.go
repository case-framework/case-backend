package messaging

import (
	"errors"
	"time"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (dbService *MessagingDBService) AddToOutgoingEmails(instanceID string, email messagingTypes.OutgoingEmail) (messagingTypes.OutgoingEmail, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if email.AddedAt <= 0 {
		email.AddedAt = time.Now().Unix()
	}

	res, err := dbService.collectionOutgoingEmails(instanceID).InsertOne(ctx, email)
	if err != nil {
		return email, err
	}
	email.ID = res.InsertedID.(primitive.ObjectID)
	return email, nil
}

func (dbService *MessagingDBService) AddToSentEmails(instanceID string, email messagingTypes.OutgoingEmail) (messagingTypes.OutgoingEmail, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()
	email.Content = ""
	email.SentAt = time.Now().UTC()
	email.To = []string{}

	email.ID = primitive.NilObjectID
	res, err := dbService.collectionSentEmails(instanceID).InsertOne(ctx, email)
	if err != nil {
		return email, err
	}
	email.ID = res.InsertedID.(primitive.ObjectID)
	return email, nil
}

func (dbService *MessagingDBService) GetSentEmailsForUser(instanceID string, userID string) (emails []messagingTypes.OutgoingEmail, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"userId": userID}
	cursor, err := dbService.collectionSentEmails(instanceID).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &emails); err != nil {
		return nil, err
	}
	return emails, nil
}

func (dbService *MessagingDBService) GetOutgoingEmailsForSending(
	instanceID string,
	lastSendAttemptOlderThan int64,
	onlyHighPrio bool,
	amount int,
) (emails []messagingTypes.OutgoingEmail, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"lastSendAttempt": bson.M{"$lt": lastSendAttemptOlderThan}}
	if onlyHighPrio {
		filter["highPrio"] = true
	}
	update := bson.M{"$set": bson.M{"lastSendAttempt": time.Now().Unix()}}

	counter := 0
	for counter < amount {
		var email messagingTypes.OutgoingEmail

		if err := dbService.collectionOutgoingEmails(instanceID).FindOneAndUpdate(
			ctx,
			filter,
			update,
		).Decode(&email); err != nil {
			break
		}
		emails = append(emails, email)
		counter++
	}

	return emails, nil
}

func (dbService *MessagingDBService) ResetLastSendAttemptForOutgoing(instanceID string, emailID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(emailID)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"lastSendAttempt": 0}}
	res, err := dbService.collectionOutgoingEmails(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount < 1 {
		return errors.New("no outgoing email found with the given id")
	}
	return nil
}

func (dbService *MessagingDBService) DeleteOutgoingEmail(instanceID string, id string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	res, err := dbService.collectionOutgoingEmails(instanceID).DeleteOne(ctx, filter, nil)
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("no outgoing email found with the given id")
	}
	return nil
}
