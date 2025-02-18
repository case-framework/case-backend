package study

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func (dbService *StudyDBService) createIndexForStudyInfosCollection(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if _, err := dbService.collectionStudyInfos(instanceID).Indexes().DropAll(ctx); err != nil {
		slog.Error("Error dropping indexes for studyInfos", slog.String("error", err.Error()))
	}

	_, err := dbService.collectionStudyInfos(instanceID).Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "key", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

// get studies
func (dbService *StudyDBService) GetStudies(instanceID string, statusFilter string, onlyKeys bool) (studies []studyTypes.Study, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{}
	if statusFilter != "" {
		filter["status"] = statusFilter
	}
	opts := options.Find()
	if onlyKeys {
		projection := bson.D{
			primitive.E{Key: "key", Value: 1},
			primitive.E{Key: "secretKey", Value: 1},
			primitive.E{Key: "configs.idMappingMethod", Value: 1},
		}
		opts.SetProjection(projection)
	}
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &studies)
	if err != nil {
		return nil, err
	}

	return studies, nil
}

func (dbService *StudyDBService) CreateStudy(instanceID string, study studyTypes.Study) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	_, err := collection.InsertOne(ctx, study)
	if err != nil {
		return err
	}

	studyKey := study.Key

	// index on surveys
	err = dbService.CreateIndexForSurveyCollection(instanceID, studyKey)
	if err != nil {
		slog.Error("Error creating index for surveys: ", slog.String("error", err.Error()))
	}

	// index on participants
	err = dbService.CreateIndexForParticipantsCollection(instanceID, studyKey)
	if err != nil {
		slog.Error("Error creating index for participants: ", slog.String("error", err.Error()))
	}

	// index on responses
	err = dbService.CreateIndexForResponsesCollection(instanceID, studyKey)
	if err != nil {
		slog.Error("Error creating index for responses: ", slog.String("error", err.Error()))
	}

	// index on reports
	err = dbService.CreateIndexForReportsCollection(instanceID, studyKey)
	if err != nil {
		slog.Error("Error creating index for reports: ", slog.String("error", err.Error()))
	}
	return nil
}

// get study by study key
func (dbService *StudyDBService) GetStudy(instanceID string, studyKey string) (study studyTypes.Study, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}
	err = collection.FindOne(ctx, filter).Decode(&study)
	if err != nil {
		return study, err
	}

	return study, nil
}

// update study status
func (dbService *StudyDBService) UpdateStudyStatus(instanceID string, studyKey string, status string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}
	update := bson.M{"$set": bson.M{"status": status}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// update study is default
func (dbService *StudyDBService) UpdateStudyIsDefault(instanceID string, studyKey string, isDefault bool) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}
	update := bson.M{"$set": bson.M{"props.systemDefaultStudy": isDefault}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (dbService *StudyDBService) UpdateStudyFileUploadRule(instanceID string, studyKey string, fileUploadRule *studyTypes.Expression) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}
	update := bson.M{"$set": bson.M{"configs.participantFileUploadRule": fileUploadRule}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (dbService *StudyDBService) UpdateStudyDisplayProps(instanceID string, studyKey string, name []studyTypes.LocalisedObject, description []studyTypes.LocalisedObject, tags []studyTypes.Tag) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}
	update := bson.M{"$set": bson.M{"props.name": name, "props.description": description, "props.tags": tags}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (dbService *StudyDBService) UpdateStudyStats(instanceID string, studyKey string, stats studyTypes.StudyStats) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"key": studyKey,
	}
	update := bson.M{"$set": bson.M{"studyStats": stats}}
	_, err := dbService.collectionStudyInfos(instanceID).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (dbService *StudyDBService) GetNotificationSubscriptions(instanceID string, studyKey string) ([]studyTypes.NotificationSubscription, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}

	var study studyTypes.Study
	err := collection.FindOne(ctx, filter).Decode(&study)
	if err != nil {
		return nil, err
	}

	return study.NotificationSubscriptions, nil
}

func (dbService *StudyDBService) UpdateStudyNotificationSubscriptions(instanceID string, studyKey string, subscriptions []studyTypes.NotificationSubscription) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}
	update := bson.M{"$set": bson.M{"notificationSubscriptions": subscriptions}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// delete study by study key
func (dbService *StudyDBService) DeleteStudy(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	// delete study collections
	err := dbService.collectionFiles(instanceID, studyKey).Drop(ctx)
	if err != nil {
		slog.Error("Error deleting collection", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	err = dbService.collectionParticipants(instanceID, studyKey).Drop(ctx)
	if err != nil {
		slog.Error("Error deleting collection", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	err = dbService.collectionReports(instanceID, studyKey).Drop(ctx)
	if err != nil {
		slog.Error("Error deleting collection", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	err = dbService.collectionResponses(instanceID, studyKey).Drop(ctx)
	if err != nil {
		slog.Error("Error deleting collection", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	err = dbService.collectionSurveys(instanceID, studyKey).Drop(ctx)
	if err != nil {
		slog.Error("Error deleting collection", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	err = dbService.collectionConfidentialResponses(instanceID, studyKey).Drop(ctx)
	if err != nil {
		slog.Error("Error deleting collection", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	err = dbService.collectionResearcherMessages(instanceID, studyKey).Drop(ctx)
	if err != nil {
		slog.Error("Error deleting collection", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	// delete study rules for study
	err = dbService.deleteStudyRules(instanceID, studyKey)
	if err != nil {
		slog.Error("Error deleting study rules", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	err = dbService.RemoveConfidentialIDMapEntriesForStudy(instanceID, studyKey)
	if err != nil {
		slog.Error("Error deleting confidential ID map entries", slog.String("studyKey", studyKey), slog.String("error", err.Error()))
	}

	collection := dbService.collectionStudyInfos(instanceID)
	filter := bson.M{"key": studyKey}
	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}
