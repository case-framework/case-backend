package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TASK_STATUS_IN_PROGRESS = "in_progress"
	TASK_STATUS_COMPLETED   = "completed"

	TASK_FILE_TYPE_JSON = "application/json"
	TASK_FILE_TYPE_CSV  = "text/csv"
)

type Task struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	CreatedBy      string             `bson:"createdBy" json:"createdBy"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
	Status         string             `bson:"status" json:"status"`
	TargetCount    int                `bson:"targetCount" json:"targetCount"`
	ProcessedCount int                `bson:"processedCount" json:"processedCount"`
	ResultFile     string             `bson:"resultFile" json:"resultFile"`
	FileType       string             `bson:"fileType" json:"fileType"`
	Error          string             `bson:"error,omitempty" json:"error,omitempty"`
}
