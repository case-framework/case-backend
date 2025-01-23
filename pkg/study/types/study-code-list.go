package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyCodeListEntry struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	StudyKey string             `bson:"studyKey" json:"studyKey"`
	ListKey  string             `bson:"listKey" json:"listKey"`
	Code     string             `bson:"code" json:"code"`
	AddedAt  time.Time          `bson:"addedAt" json:"addedAt"`
}
