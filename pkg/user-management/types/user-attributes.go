package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAttributes struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	Type       string             `bson:"type" json:"type"`
	Attributes map[string]any     `bson:"attributes" json:"attributes"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}
