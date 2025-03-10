package messaging

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

func (messagingDBService *MessagingDBService) CreateIndexForSMSTemplates(instanceID string) error {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	_, err := messagingDBService.collectionSMSTemplates(instanceID).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "messageType", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

// save email template (if id is empty, insert, else update)
func (messagingDBService *MessagingDBService) SaveSMSTemplate(instanceID string, smsTemplate messagingTypes.SMSTemplate) (messagingTypes.SMSTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	if smsTemplate.ID.IsZero() {
		smsTemplate.ID = primitive.NewObjectID()
		// new template
		res, err := messagingDBService.collectionSMSTemplates(instanceID).InsertOne(ctx, smsTemplate)
		if err != nil {
			return messagingTypes.SMSTemplate{}, err
		}
		smsTemplate.ID = res.InsertedID.(primitive.ObjectID)
		return smsTemplate, nil
	}

	// update template
	filter := bson.M{"_id": smsTemplate.ID}
	upsert := false
	after := options.After
	opt := options.FindOneAndReplaceOptions{Upsert: &upsert, ReturnDocument: &after}
	err := messagingDBService.collectionSMSTemplates(instanceID).FindOneAndReplace(ctx, filter, smsTemplate, &opt).Decode(&smsTemplate)
	if err != nil {
		return messagingTypes.SMSTemplate{}, err
	}
	return smsTemplate, nil
}

func (messagingDBService *MessagingDBService) GetSMSTemplateByType(instanceID string, messageType string) (*messagingTypes.SMSTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"messageType": messageType}

	var smsTemplate messagingTypes.SMSTemplate
	err := messagingDBService.collectionSMSTemplates(instanceID).FindOne(ctx, filter).Decode(&smsTemplate)
	if err != nil {
		return nil, err
	}
	return &smsTemplate, nil
}
