package participantuser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TempToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"token_id,omitempty"`
	Token      string             `bson:"token" json:"token"`
	Expiration time.Time          `bson:"expiration" json:"expiration"`
	Purpose    string             `bson:"purpose" json:"purpose"`
	UserID     string             `bson:"userID" json:"userID"`
	Info       map[string]string  `bson:"info" json:"info"`
	InstanceID string             `bson:"instanceID" json:"instanceID"`
}
