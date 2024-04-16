package messaging

import (
	study "github.com/case-framework/case-backend/pkg/study/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScheduledEmail struct {
	ID        primitive.ObjectID   `bson:"_id" json:"id,omitempty"`
	Template  EmailTemplate        `bson:"template" json:"template"`
	Type      string               `bson:"type" json:"type"`
	StudyKey  string               `bson:"studyKey" json:"studyKey"`
	Condition *study.ExpressionArg `bson:"condition" json:"condition"`
	NextTime  int64                `bson:"nextTime" json:"nextTime"`
	Period    int64                `bson:"period" json:"period"`
	Label     string               `bson:"label" json:"label"`
	Until     int64                `bson:"until" json:"until"`
}
