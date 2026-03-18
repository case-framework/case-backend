package study

import (
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

var studyRuleIndexNames []string

var indexesForStudyRulesCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "studyKey", Value: 1},
		},
		Options: options.Index().SetName("studyKey_1"),
	},
	{
		Keys: bson.D{
			{Key: "uploadedAt", Value: 1},
			{Key: "studyKey", Value: 1},
		},
		Options: options.Index().SetName("uploadedAt_1_studyKey_1"),
	},
}

func (dbService *StudyDBService) DropIndexForStudyRulesCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)

	if dropAll {
		err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for studyRules", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, indexName := range studyRuleIndexNames {
			if indexName == "" {
				slog.Error("Index name is empty for studyRules collection", slog.String("index", fmt.Sprintf("%+v", indexName)))
				continue
			}
			err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for studyRules", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForStudyRulesCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)
	names, err := collection.Indexes().CreateMany(ctx, indexesForStudyRulesCollection)
	if err != nil {
		slog.Error("Error creating index for studyRules", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
	studyRuleIndexNames = names
}

func (dbService *StudyDBService) deleteStudyRules(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)
	_, err := collection.DeleteMany(ctx, bson.M{"studyKey": studyKey})
	return err
}

func (dbService *StudyDBService) SaveStudyRules(instanceID string, studyKey string, rules studyTypes.StudyRules) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)
	_, err := collection.InsertOne(ctx, rules)
	return err
}

func (dbService *StudyDBService) GetCurrentStudyRules(instanceID string, studyKey string) (rules studyTypes.StudyRules, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)

	sortByPublished := bson.D{
		{Key: "uploadedAt", Value: -1},
	}

	filter := bson.M{
		"studyKey": studyKey,
	}

	opts := options.FindOne().SetSort(sortByPublished)

	err = collection.FindOne(ctx, filter, opts).Decode(&rules)
	if err != nil {
		return rules, err
	}
	err = rules.UnmarshalRules()
	return rules, err
}

func (dbService *StudyDBService) GetStudyRulesByID(instanceID string, studyKey string, id string) (rules studyTypes.StudyRules, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)

	_id, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return rules, err
	}

	filter := bson.M{
		"studyKey": studyKey,
		"_id":      _id,
	}

	err = collection.FindOne(ctx, filter).Decode(&rules)
	if err != nil {
		return rules, err
	}
	err = rules.UnmarshalRules()

	return rules, err
}

func (dbService *StudyDBService) DeleteStudyRulesByID(instanceID string, studyKey string, id string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)

	_id, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"studyKey": studyKey,
		"_id":      _id,
	}

	res, err := collection.DeleteOne(ctx, filter)
	if res.DeletedCount < 1 {
		return mongo.ErrNoDocuments
	}
	return err
}

func (dbService *StudyDBService) GetStudyRulesHistory(instanceID string, studyKey string) (ruleHistory []studyTypes.StudyRules, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)

	filter := bson.M{
		"studyKey": studyKey,
	}

	opts := options.Find().SetSort(bson.D{{Key: "uploadedAt", Value: -1}})
	opts.SetProjection(bson.D{
		{Key: "rules", Value: 0},
		{Key: "serialisedRules", Value: 0},
	})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return ruleHistory, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &ruleHistory); err != nil {
		return ruleHistory, err
	}
	return ruleHistory, nil
}
