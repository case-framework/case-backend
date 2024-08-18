package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SentSMS struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"userID" json:"userID"`
	MessageType string             `bson:"messageType" json:"messageType"`
	SentAt      time.Time          `bson:"sentAt" json:"sentAt"`
	PhoneNumber string             `bson:"phoneNumber" json:"phoneNumber"`
}

type SMSTemplate struct {
	ID              primitive.ObjectID  `bson:"_id" json:"id,omitempty"`
	MessageType     string              `bson:"messageType" json:"messageType"`
	DefaultLanguage string              `bson:"defaultLanguage" json:"defaultLanguage"`
	From            string              `bson:"from" json:"from"`
	Translations    []LocalizedTemplate `bson:"translations" json:"translations"`
}
