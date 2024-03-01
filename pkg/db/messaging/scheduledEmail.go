package messaging

import (
	messagingTypes "github.com/case-framework/case-backend/pkg/types/messaging"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// get all scheduled emails
func (dbService *MessagingDBService) GetAllScheduledEmails(instanceID string) ([]messagingTypes.ScheduledEmail, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}

	collection := dbService.collectionEmailSchedules(instanceID)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var scheduledEmails []messagingTypes.ScheduledEmail
	if err = cursor.All(ctx, &scheduledEmails); err != nil {
		return nil, err
	}

	return scheduledEmails, nil
}

// get scheduled email by id
func (dbService *MessagingDBService) GetScheduledEmailByID(instanceID string, id string) (*messagingTypes.ScheduledEmail, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"_id": id}

	var scheduledEmail messagingTypes.ScheduledEmail
	err := dbService.collectionEmailSchedules(instanceID).FindOne(ctx, filter).Decode(&scheduledEmail)
	if err != nil {
		return nil, err
	}
	return &scheduledEmail, nil
}

// save scheduled email
func (dbService *MessagingDBService) SaveScheduledEmail(instanceID string, scheduledEmail messagingTypes.ScheduledEmail) (messagingTypes.ScheduledEmail, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if !scheduledEmail.ID.IsZero() {
		filter := bson.M{"_id": scheduledEmail.ID}

		upsert := false
		rd := options.After
		options := options.FindOneAndReplaceOptions{
			Upsert:         &upsert,
			ReturnDocument: &rd,
		}
		elem := messagingTypes.ScheduledEmail{}
		err := dbService.collectionEmailSchedules(instanceID).FindOneAndReplace(
			ctx, filter, scheduledEmail, &options,
		).Decode(&elem)
		return elem, err
	} else {
		scheduledEmail.ID = primitive.NewObjectID()
		res, err := dbService.collectionEmailSchedules(instanceID).InsertOne(ctx, scheduledEmail)
		if err != nil {
			return scheduledEmail, err
		}
		scheduledEmail.ID = res.InsertedID.(primitive.ObjectID)
		return scheduledEmail, nil
	}
}

// delete scheduled email
func (dbService *MessagingDBService) DeleteScheduledEmail(instanceID string, id string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"_id": id}

	_, err := dbService.collectionEmailSchedules(instanceID).DeleteOne(ctx, filter)
	return err
}
