package study

import (
	"fmt"
	"log/slog"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studytypes "github.com/case-framework/case-backend/pkg/study/types"
)

var indexesForStudyVariablesCollection = []mongo.IndexModel{
	{
		Keys: bson.D{
			{Key: "studyKey", Value: 1},
			{Key: "key", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("studyKey_1_key_1"),
	},
}

func getValueProjection() bson.D {
	return bson.D{
		{Key: "type", Value: 1},
		{Key: "value", Value: 1},
		{Key: "studyKey", Value: 1},
		{Key: "key", Value: 1},
		{Key: "valueUpdatedAt", Value: 1},
	}
}

func (dbService *StudyDBService) DropIndexForStudyVariablesCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	collection := dbService.collectionStudyVariables(instanceID)
	if dropAll {
		_, err := collection.Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for studyVariables", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, index := range indexesForStudyVariablesCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for studyVariables collection", slog.String("index", fmt.Sprintf("%+v", index)), slog.String("instanceID", instanceID))
				continue
			}
			indexName := *index.Options.Name
			_, err := collection.Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for studyVariables", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForStudyVariablesCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionStudyVariables(instanceID).Indexes().CreateMany(ctx, indexesForStudyVariablesCollection)
	if err != nil {
		slog.Error("Error creating index for studyVariables", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

// create a study variable
func (dbService *StudyDBService) CreateStudyVariable(instanceID string, variable studytypes.StudyVariables) (string, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	// Ensure timestamps
	now := time.Now().UTC()
	if variable.CreatedAt.IsZero() {
		variable.CreatedAt = now
	}
	if variable.ConfigUpdatedAt.IsZero() {
		variable.ConfigUpdatedAt = now
	}
	if variable.ValueUpdatedAt.IsZero() {
		variable.ValueUpdatedAt = now
	}

	res, err := dbService.collectionStudyVariables(instanceID).InsertOne(ctx, variable)
	if err != nil {
		return "", err
	}
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), nil
}

// update a study variable's config by id
func (dbService *StudyDBService) UpdateStudyVariableConfig(
	instanceID string,
	studyKey string,
	key string,
	label string,
	description string,
	uiType string,
	uiPriority int,
	configs any) (studytypes.StudyVariables, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": studyKey, "key": key}

	update := bson.M{"$set": bson.M{
		"label":           label,
		"description":     description,
		"uiType":          uiType,
		"uiPriority":      uiPriority,
		"configs":         configs,
		"configUpdatedAt": time.Now().UTC(),
	}}

	var updated studytypes.StudyVariables
	err := dbService.collectionStudyVariables(instanceID).FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	return updated, err
}

// update a study variable's value by id
func (dbService *StudyDBService) UpdateStudyVariableValue(
	instanceID string,
	studyKey string,
	key string,
	value any) (studytypes.StudyVariables, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": studyKey, "key": key}

	update := bson.M{"$set": bson.M{
		"value":          value,
		"valueUpdatedAt": time.Now().UTC(),
	}}

	var updated studytypes.StudyVariables
	err := dbService.collectionStudyVariables(instanceID).FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	return updated, err
}

// get all study variables by studyKey (optionally only core fields)
func (dbService *StudyDBService) GetStudyVariablesByStudyKey(instanceID string, studyKey string, onlyValue bool) ([]studytypes.StudyVariables, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": studyKey}
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	if onlyValue {
		opts.SetProjection(getValueProjection())
	}

	cursor, err := dbService.collectionStudyVariables(instanceID).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vars []studytypes.StudyVariables
	if err := cursor.All(ctx, &vars); err != nil {
		return nil, err
	}
	return vars, nil
}

// get a study variable by id
func (dbService *StudyDBService) GetStudyVariableByID(instanceID string, id string, onlyValue bool) (studytypes.StudyVariables, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return studytypes.StudyVariables{}, err
	}

	var v studytypes.StudyVariables
	findOneOpts := options.FindOne()
	if onlyValue {
		findOneOpts.SetProjection(getValueProjection())
	}
	err = dbService.collectionStudyVariables(instanceID).FindOne(ctx, bson.M{"_id": _id}, findOneOpts).Decode(&v)
	return v, err
}

// get a study variable by studyKey and key
func (dbService *StudyDBService) GetStudyVariableByStudyKeyAndKey(instanceID string, studyKey string, key string, onlyValue bool) (studytypes.StudyVariables, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var v studytypes.StudyVariables
	findOneOpts := options.FindOne()
	if onlyValue {
		findOneOpts.SetProjection(getValueProjection())
	}
	err := dbService.collectionStudyVariables(instanceID).FindOne(ctx, bson.M{"studyKey": studyKey, "key": key}, findOneOpts).Decode(&v)
	return v, err
}

// remove a study variable by id
func (dbService *StudyDBService) DeleteStudyVariableByID(instanceID string, id string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	res, err := dbService.collectionStudyVariables(instanceID).DeleteOne(ctx, bson.M{"_id": _id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// delete a study variable by studyKey and key
func (dbService *StudyDBService) DeleteStudyVariableByStudyKeyAndKey(instanceID string, studyKey string, key string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{"studyKey": studyKey, "key": key}
	res, err := dbService.collectionStudyVariables(instanceID).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// remove all study variables by studyKey
func (dbService *StudyDBService) DeleteStudyVariablesByStudyKey(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionStudyVariables(instanceID).DeleteMany(ctx, bson.M{"studyKey": studyKey})
	return err
}
