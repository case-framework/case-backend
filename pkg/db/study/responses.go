package study

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

func (dbService *StudyDBService) CreateIndexForResponsesCollection(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionResponses(instanceID, studyKey)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "participantID", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "participantID", Value: 1},
				{Key: "key", Value: 1},
				{Key: "submittedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "submittedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "arrivedAt", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "key", Value: 1},
			},
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// get response by id
func (dbService *StudyDBService) GetResponseByID(instanceID string, studyKey string, responseID string) (response studyTypes.SurveyResponse, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(responseID)
	if err != nil {
		return response, err
	}

	filter := bson.M{
		"_id": _id,
	}

	err = dbService.collectionResponses(instanceID, studyKey).FindOne(ctx, filter).Decode(&response)
	return response, err
}

// get paginated responses by query
func (dbService *StudyDBService) GetResponses(instanceID string, studyKey string, filter bson.M, sort bson.M, page int64, limit int64) (responses []studyTypes.SurveyResponse, paginationInfo *PaginationInfos, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	totalCount, err := dbService.GetResponsesCount(instanceID, studyKey, filter)
	if err != nil {
		return responses, nil, err
	}

	paginationInfo = prepPaginationInfos(
		totalCount,
		page,
		limit,
	)

	skip := (paginationInfo.CurrentPage - 1) * paginationInfo.PageSize

	opts := options.Find().SetSort(sort).SetSkip(skip).SetLimit(paginationInfo.PageSize)
	collection := dbService.collectionResponses(instanceID, studyKey)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return responses, nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &responses)
	if err != nil {
		return responses, nil, err
	}

	return responses, paginationInfo, nil
}

// get responses count by query
func (dbService *StudyDBService) GetResponsesCount(instanceID string, studyKey string, filter bson.M) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	return dbService.collectionResponses(instanceID, studyKey).CountDocuments(ctx, filter)
}

// execute on responses by query
func (dbService *StudyDBService) FindAndExecuteOnResponses(
	ctx context.Context,
	instanceID string, studyKey string,
	filter bson.M,
	sort bson.M,
	returnOnError bool,
	fn func(dbService *StudyDBService, r studyTypes.SurveyResponse, instanceID string, studyKey string, args ...interface{}) error,
	args ...interface{},
) error {
	opts := options.Find().SetSort(sort)

	cursor, err := dbService.collectionResponses(instanceID, studyKey).Find(ctx, filter, opts)
	if err != nil {
		return err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var response studyTypes.SurveyResponse
		if err = cursor.Decode(&response); err != nil {
			slog.Error("Error while decoding response", slog.String("error", err.Error()))
			continue
		}

		if err = fn(dbService, response, instanceID, studyKey, args...); err != nil {
			slog.Error("Error while executing function on response", slog.String("responseID", response.ID.Hex()), slog.String("error", err.Error()))
			if returnOnError {
				return err
			}
			continue
		}
	}
	return nil
}

// delete response by id
func (dbService *StudyDBService) DeleteResponseByID(instanceID string, studyKey string, responseID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(responseID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}

	res, err := dbService.collectionResponses(instanceID, studyKey).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return err
}

// delete responses by query
func (dbService *StudyDBService) DeleteResponses(instanceID string, studyKey string, filter bson.M) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionResponses(instanceID, studyKey).DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return err
}
