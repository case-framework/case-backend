package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Report struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key           string             `bson:"key" json:"key"`
	ParticipantID string             `bson:"participantID" json:"participantID"` // reference to the study specific participant ID
	ResponseID    string             `bson:"responseID" json:"responseID"`       // reference to the report
	Timestamp     int64              `bson:"timestamp" json:"timestamp"`
	ModifiedAt    time.Time          `bson:"modifiedAt,omitempty" json:"modifiedAt,omitempty"` // if report is updated later, this is the time of the update
	Data          []ReportData       `bson:"data" json:"data,omitempty"`
}

type ReportData struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
	Dtype string `bson:"dtype,omitempty" json:"dtype,omitempty"`
}
