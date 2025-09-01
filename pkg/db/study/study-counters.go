package study

import (
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StudyCounter struct {
	StudyKey string `bson:"studyKey"`
	Scope    string `bson:"scope"`
	Value    int64  `bson:"value"`
}

func (dbService *StudyDBService) CreateIndexForStudyCountersCollection(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if _, err := dbService.collectionStudyCounters(instanceID).Indexes().DropAll(ctx); err != nil {
		slog.Error("Error dropping indexes for studyCounters", slog.String("error", err.Error()))
	}

	collection := dbService.collectionStudyCounters(instanceID)
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "studyKey", Value: 1},
				{Key: "scope", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Get current counter value (without incrementing)
func (dbService *StudyDBService) GetCurrentStudyCounterValue(instanceID string, studyKey string, scope string) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	counter := StudyCounter{}
	err := dbService.collectionStudyCounters(instanceID).FindOne(ctx, bson.M{"studyKey": studyKey, "scope": scope}).Decode(&counter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, err
	}

	return counter.Value, nil
}

// Get all counter values for a study
func (dbService *StudyDBService) GetAllStudyCounterValues(instanceID string, studyKey string) ([]StudyCounter, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	cursor, err := dbService.collectionStudyCounters(instanceID).Find(ctx, bson.M{"studyKey": studyKey})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var counters []StudyCounter
	if err = cursor.All(ctx, &counters); err != nil {
		return nil, err
	}
	return counters, nil
}

// Increment counter value (atomical find and update)
func (dbService *StudyDBService) IncrementAndGetStudyCounterValue(instanceID string, studyKey string, scope string) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	counter := StudyCounter{}
	err := dbService.collectionStudyCounters(instanceID).FindOneAndUpdate(
		ctx,
		bson.M{"studyKey": studyKey, "scope": scope},
		bson.M{
			"$inc":         bson.M{"value": 1},
			"$setOnInsert": bson.M{"studyKey": studyKey, "scope": scope, "value": 0},
		},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&counter)
	if err != nil {
		return 0, err
	}

	return counter.Value, nil
}

// Remove study counter value (reset to 0)
func (dbService *StudyDBService) RemoveStudyCounterValue(instanceID string, studyKey string, scope string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionStudyCounters(instanceID).DeleteOne(ctx, bson.M{"studyKey": studyKey, "scope": scope})
	return err
}

// Remove all study counters for a study
func (dbService *StudyDBService) RemoveAllStudyCounters(instanceID string, studyKey string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionStudyCounters(instanceID).DeleteMany(ctx, bson.M{"studyKey": studyKey})
	return err
}
