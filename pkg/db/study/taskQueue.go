package study

import (
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	studyTypes "github.com/case-framework/case-backend/pkg/study/types"
)

const (
	REMOVE_TASK_FROM_QUEUE_AFTER = 60 * 60 * 24 * 2 // 2 days
)

var indexesForTaskQueueCollection = []mongo.IndexModel{
	{
		Keys:    bson.D{{Key: "updatedAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(REMOVE_TASK_FROM_QUEUE_AFTER).SetName("updatedAt_1"),
	},
}

func (dbService *StudyDBService) DropIndexForTaskQueueCollection(instanceID string, dropAll bool) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	if dropAll {
		_, err := dbService.collectionTaskQueue(instanceID).Indexes().DropAll(ctx)
		if err != nil {
			slog.Error("Error dropping all indexes for task queue", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
		}
	} else {
		for _, index := range indexesForTaskQueueCollection {
			if index.Options == nil || index.Options.Name == nil {
				slog.Error("Index name is nil for task queue collection", slog.String("index", fmt.Sprintf("%+v", index)))
				continue
			}
			indexName := *index.Options.Name
			_, err := dbService.collectionTaskQueue(instanceID).Indexes().DropOne(ctx, indexName)
			if err != nil {
				slog.Error("Error dropping index for task queue", slog.String("error", err.Error()), slog.String("instanceID", instanceID), slog.String("indexName", indexName))
			}
		}
	}
}

func (dbService *StudyDBService) CreateDefaultIndexesForTaskQueueCollection(instanceID string) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionTaskQueue(instanceID).Indexes().CreateMany(ctx, indexesForTaskQueueCollection)
	if err != nil {
		slog.Error("Error creating index for task queue", slog.String("error", err.Error()), slog.String("instanceID", instanceID))
	}
}

// create task
func (dbService *StudyDBService) CreateTask(
	instanceID string,
	createdBy string,
	targetCount int,
	fileType string,
) (task studyTypes.Task, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	task = studyTypes.Task{
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Status:         studyTypes.TASK_STATUS_IN_PROGRESS,
		TargetCount:    targetCount,
		ProcessedCount: 0,
		FileType:       fileType,
	}

	ret, err := dbService.collectionTaskQueue(instanceID).InsertOne(ctx, task)
	if err != nil {
		return task, err
	}
	task.ID = ret.InsertedID.(primitive.ObjectID)
	return task, nil
}

// get task by id
func (dbService *StudyDBService) GetTaskByID(instanceID string, taskID string) (task studyTypes.Task, err error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return task, err
	}

	filter := bson.M{
		"_id": _id,
	}

	err = dbService.collectionTaskQueue(instanceID).FindOne(ctx, filter).Decode(&task)
	return task, err
}

func (dbService *StudyDBService) GetTaskByFilename(instanceID string, filename string) (studyTypes.Task, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	var task studyTypes.Task

	filter := bson.M{
		"resultFile": filename,
	}

	err := dbService.collectionTaskQueue(instanceID).FindOne(ctx, filter).Decode(&task)
	if err != nil {
		return task, err
	}
	return task, nil
}

func (dbService *StudyDBService) UpdateTaskTotalCount(instanceID string, taskID string, totalCount int) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}
	update := bson.M{
		"$set": bson.M{
			"targetCount": totalCount,
			"updatedAt":   time.Now(),
		},
	}
	_, err = dbService.collectionTaskQueue(instanceID).UpdateOne(ctx, filter, update)
	return err
}

// update task processed count
func (dbService *StudyDBService) UpdateTaskProgress(instanceID string, taskID string, processedCount int) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}
	update := bson.M{
		"$set": bson.M{
			"processedCount": processedCount,
			"updatedAt":      time.Now(),
		},
	}
	_, err = dbService.collectionTaskQueue(instanceID).UpdateOne(ctx, filter, update)
	return err
}

func (dbService *StudyDBService) UpdateTaskCompleted(
	instanceID string,
	taskID string,
	status string,
	processedCount int,
	errMsg string,
	resultFile string,
) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}
	update := bson.M{
		"$set": bson.M{
			"processedCount": processedCount,
			"status":         status,
			"error":          errMsg,
			"resultFile":     resultFile,
			"updatedAt":      time.Now(),
		},
	}
	_, err = dbService.collectionTaskQueue(instanceID).UpdateOne(ctx, filter, update)
	return err
}

// delete task by id
func (dbService *StudyDBService) DeleteTaskByID(instanceID string, taskID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": _id,
	}
	res, err := dbService.collectionTaskQueue(instanceID).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
