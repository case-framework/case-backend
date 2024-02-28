package messaging

import "go.mongodb.org/mongo-driver/bson"

// find all email templates with study key empty
func (messagingDBService *MessagingDBService) GetGlobalEmailTemplates(instanceID string) ([]EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": bson.M{"$exists": false}}

	var emailTemplates []EmailTemplate
	cursor, err := messagingDBService.collectionEmailTemplates(instanceID).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &emailTemplates); err != nil {
		return nil, err
	}
	return emailTemplates, nil
}

// find one email template by message type and study key empty
func (messagingDBService *MessagingDBService) GetGlobalEmailTemplateByMessageType(instanceID string, messageType string) (*EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"messageType": messageType, "studyKey": bson.M{"$exists": false}}

	var emailTemplate EmailTemplate
	err := messagingDBService.collectionEmailTemplates(instanceID).FindOne(ctx, filter).Decode(&emailTemplate)
	if err != nil {
		return nil, err
	}
	return &emailTemplate, nil
}

// find one email template by id

// create a new email template

// update an email template

// delete an email template by message type and study key empty

// find all email templates with study key non-empty

// find one email template by message type and study key
