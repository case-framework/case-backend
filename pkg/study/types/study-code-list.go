package types

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type StudyCodeListEntry struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	StudyKey string        `bson:"studyKey" json:"studyKey"`
	ListKey  string        `bson:"listKey" json:"listKey"`
	Code     string        `bson:"code" json:"code"`
	AddedAt  time.Time     `bson:"addedAt" json:"addedAt"`
}
