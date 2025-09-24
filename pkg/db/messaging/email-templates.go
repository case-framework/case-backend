package messaging

import (
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	messagingTypes "github.com/case-framework/case-backend/pkg/messaging/types"
)

var indexesForEmailTemplatesCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "messageType", Value: 1},
			{Key: "studyKey", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("messageType_studyKey_1"),
	},
}

func (messagingDBService *MessagingDBService) DropIndexForEmailTemplatesCollection(instanceID string, dropAll bool) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	if dropAll {
		_, err := messagingDBService.collectionEmailTemplates(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for email templates", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, index := range indexesForEmailTemplatesCollection {
			if index.Options.Name == nil {
				slog.Error("Index name is nil for email templates collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := messagingDBService.collectionEmailTemplates(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for email templates", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (messagingDBService *MessagingDBService) CreateDefaultIndexesForEmailTemplatesCollection(instanceID string) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()
	_, err := messagingDBService.collectionEmailTemplates(instanceID).Indexes().CreateMany(ctx, indexesForEmailTemplatesCollection)
	if err != nil {
		slog.Error("Error creating index for email templates", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

// find all email templates with study key empty
func (messagingDBService *MessagingDBService) GetGlobalEmailTemplates(instanceID string) ([]messagingTypes.EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": bson.M{"$exists": false}}

	var emailTemplates []messagingTypes.EmailTemplate
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
func (messagingDBService *MessagingDBService) GetGlobalEmailTemplateByMessageType(instanceID string, messageType string) (*messagingTypes.EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"messageType": messageType, "studyKey": bson.M{"$exists": false}}

	var emailTemplate messagingTypes.EmailTemplate
	err := messagingDBService.collectionEmailTemplates(instanceID).FindOne(ctx, filter).Decode(&emailTemplate)
	if err != nil {
		return nil, err
	}
	return &emailTemplate, nil
}

// find one email template by id
func (messagingDBService *MessagingDBService) GetEmailTemplateByID(instanceID string, id string) (*messagingTypes.EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": _id}

	var emailTemplate messagingTypes.EmailTemplate
	err = messagingDBService.collectionEmailTemplates(instanceID).FindOne(ctx, filter).Decode(&emailTemplate)
	if err != nil {
		return nil, err
	}
	return &emailTemplate, nil
}

// save email template (if id is empty, insert, else update)
func (messagingDBService *MessagingDBService) SaveEmailTemplate(instanceID string, emailTemplate messagingTypes.EmailTemplate) (messagingTypes.EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	if emailTemplate.ID.IsZero() {
		emailTemplate.ID = primitive.NewObjectID()
		// new email template
		res, err := messagingDBService.collectionEmailTemplates(instanceID).InsertOne(ctx, emailTemplate)
		if err != nil {
			return messagingTypes.EmailTemplate{}, err
		}
		emailTemplate.ID = res.InsertedID.(primitive.ObjectID)
		return emailTemplate, nil
	}

	// update email template
	filter := bson.M{"_id": emailTemplate.ID}
	upsert := false
	after := options.After
	opt := options.FindOneAndReplaceOptions{Upsert: &upsert, ReturnDocument: &after}
	err := messagingDBService.collectionEmailTemplates(instanceID).FindOneAndReplace(ctx, filter, emailTemplate, &opt).Decode(&emailTemplate)
	if err != nil {
		return messagingTypes.EmailTemplate{}, err
	}
	return emailTemplate, nil
}

// delete an email template by message type and study key
func (messagingDBService *MessagingDBService) DeleteEmailTemplate(instanceID string, messageType string, studyKey string) error {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"messageType": messageType, "studyKey": studyKey}
	if studyKey == "" {
		filter["studyKey"] = bson.M{"$exists": false}
	}
	_, err := messagingDBService.collectionEmailTemplates(instanceID).DeleteOne(ctx, filter)
	return err
}

// find all email templates with study key non-empty
func (messagingDBService *MessagingDBService) GetEmailTemplatesForAllStudies(instanceID string) ([]messagingTypes.EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": bson.M{"$exists": true}}

	var emailTemplates []messagingTypes.EmailTemplate
	cursor, err := messagingDBService.collectionEmailTemplates(instanceID).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &emailTemplates); err != nil {
		return nil, err
	}
	return emailTemplates, nil
}

// find all email templates by study key
func (messagingDBService *MessagingDBService) GetStudyEmailTemplates(instanceID string, studyKey string) ([]messagingTypes.EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": studyKey}

	var emailTemplates []messagingTypes.EmailTemplate
	cursor, err := messagingDBService.collectionEmailTemplates(instanceID).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &emailTemplates); err != nil {
		return nil, err
	}
	return emailTemplates, nil
}

// find one email template by message type and study key
func (messagingDBService *MessagingDBService) GetStudyEmailTemplateByMessageType(instanceID string, studyKey string, messageType string) (*messagingTypes.EmailTemplate, error) {
	ctx, cancel := messagingDBService.getContext()
	defer cancel()

	filter := bson.M{"messageType": messageType, "studyKey": studyKey}

	var emailTemplate messagingTypes.EmailTemplate
	err := messagingDBService.collectionEmailTemplates(instanceID).FindOne(ctx, filter).Decode(&emailTemplate)
	if err != nil {
		return nil, err
	}
	return &emailTemplate, nil
}
