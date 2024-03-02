package study

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

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
