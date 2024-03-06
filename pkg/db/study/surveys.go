package study

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

func (dbService *StudyDBService) CreateIndexForSurveyCollection(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionSurveys(instanceID, studyKey)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "surveyDefinition.key", Value: 1},
				{Key: "unpublished", Value: 1},
				{Key: "published", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "published", Value: 1},
				{Key: "surveyDefinition.key", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "unpublished", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "surveyDefinition.key", Value: 1},
				{Key: "versionID", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (dbService *StudyDBService) SaveSurveyVersion(instanceID string, studyKey string, survey *studyTypes.Survey) (err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	ret, err := dbService.collectionSurveys(instanceID, studyKey).InsertOne(ctx, survey)
	if err != nil {
		return err
	}
	survey.ID = ret.InsertedID.(primitive.ObjectID)

	return nil
}

func (dbService *StudyDBService) GetSurveyKeysForStudy(instanceID string, studyKey string, includeUnpublished bool) (surveyKeys []string, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if !includeUnpublished {
		filter["unpublished"] = 0
	}
	res, err := dbService.collectionSurveys(instanceID, studyKey).Distinct(ctx, "surveyDefinition.key", filter)
	if err != nil {
		return surveyKeys, err
	}
	surveyKeys = make([]string, len(res))
	for i, r := range res {
		surveyKeys[i] = r.(string)
	}
	return surveyKeys, err
}

var (
	sortByPublishedDesc = bson.D{
		primitive.E{Key: "published", Value: -1},
	}

	projectionToRemoveSurveyContentAndRules = bson.D{
		primitive.E{Key: "surveyDefinition.items", Value: 0},
		primitive.E{Key: "prefillRules", Value: 0},
		primitive.E{Key: "contextRules", Value: 0},
	}
)

func (dbService *StudyDBService) GetSurveyVersions(instanceID string, studyKey string, surveyKey string) (surveys []*studyTypes.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{}
	if len(surveyKey) > 0 {
		filter["surveyDefinition.key"] = surveyKey
	}
	opts := &options.FindOptions{}

	opts.SetProjection(projectionToRemoveSurveyContentAndRules)

	opts.SetSort(sortByPublishedDesc)

	cur, err := dbService.collectionSurveys(instanceID, studyKey).Find(
		ctx,
		filter,
		opts,
	)
	if err != nil {
		return surveys, err
	}

	if err = cur.All(ctx, &surveys); err != nil {
		return nil, err
	}
	return surveys, nil
}

func (dbService *StudyDBService) GetSurveyVersion(instanceID string, studyKey string, surveyKey string, versionID string) (survey *studyTypes.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"versionID":            versionID,
	}

	err = dbService.collectionSurveys(instanceID, studyKey).FindOne(ctx, filter).Decode(&survey)
	if err != nil {
		return nil, err
	}
	return survey, nil
}

func (dbService *StudyDBService) GetCurrentSurveyVersion(instanceID string, studyKey string, surveyKey string) (survey *studyTypes.Survey, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"unpublished":          0,
	}

	opts := &options.FindOneOptions{}
	opts.SetSort(sortByPublishedDesc)

	err = dbService.collectionSurveys(instanceID, studyKey).FindOne(ctx, filter, opts).Decode(&survey)
	if err != nil {
		return nil, err
	}
	return survey, nil
}

func (dbService *StudyDBService) DeleteSurveyVersion(instanceID string, studyKey string, surveyKey string, versionID string) (err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"versionID":            versionID,
	}

	res, err := dbService.collectionSurveys(instanceID, studyKey).DeleteOne(ctx, filter)
	if res.DeletedCount < 1 {
		return errors.New("no item was deleted")
	}
	return err
}

func (dbService *StudyDBService) UnpublishSurvey(instanceID string, studyKey string, surveyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"surveyDefinition.key": surveyKey,
		"unpublished":          bson.M{"$not": bson.M{"$gt": 0}},
	}
	update := bson.M{"$set": bson.M{"unpublished": time.Now().Unix()}}
	_, err := dbService.collectionSurveys(instanceID, studyKey).UpdateMany(ctx, filter, update)
	return err
}
