package study

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

func (dbService *StudyDBService) CreateIndexForStudyRulesCollection(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyRules(instanceID)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "studyKey", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "uploadedAt", Value: 1},
				{Key: "studyKey", Value: 1},
			},
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
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
		primitive.E{Key: "uploadedAt", Value: -1},
	}

	filter := bson.M{
		"studyKey": studyKey,
	}

	opts := &options.FindOneOptions{
		Sort: sortByPublished,
	}

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

	_id, err := primitive.ObjectIDFromHex(id)
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

	_id, err := primitive.ObjectIDFromHex(id)
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
		primitive.E{Key: "rules", Value: 0},
		primitive.E{Key: "serialisedRules", Value: 0},
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
