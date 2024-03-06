package study

import "go.mongodb.org/mongo-driver/bson/primitive"

type StudyRules struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	StudyKey   string             `bson:"studyKey" json:"studyKey"`
	UploadedAt int64              `bson:"uploadedAt" json:"uploadedAt"`
	UploadedBy string             `bson:"uploadedBy" json:"uploadedBy"`
	Rules      []Expression       `bson:"rules" json:"rules"`
}
