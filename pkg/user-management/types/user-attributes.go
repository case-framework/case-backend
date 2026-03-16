package types

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserAttributes struct {
	ID         bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID     bson.ObjectID  `bson:"userId" json:"userId"`
	Type       string         `bson:"type" json:"type"`
	Attributes map[string]any `bson:"attributes" json:"attributes"`
	CreatedAt  time.Time      `bson:"createdAt" json:"createdAt"`
}
