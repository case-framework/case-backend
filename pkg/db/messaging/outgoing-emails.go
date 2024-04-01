package messaging

import (
	"time"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
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
	email.AddedAt = time.Now().Unix()
	email.Content = ""

	email.ID = primitive.NilObjectID
	res, err := dbService.collectionSentEmails(instanceID).InsertOne(ctx, email)
	if err != nil {
		return email, err
	}
	email.ID = res.InsertedID.(primitive.ObjectID)
	return email, nil
}
