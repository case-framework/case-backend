package study

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	studyTypes "github.com/case-framework/case-backend/pkg/types/study"
)

// create task
func (dbService *StudyDBService) CreateTask(
	instanceID string,
	createdBy string,
	targetCount int,
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
