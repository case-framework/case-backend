package types

import "time"

type SentSMS struct {
	ID          string    `bson:"_id" json:"id"`
	UserID      string    `bson:"userID" json:"userID"`
	MessageType string    `bson:"messageType" json:"messageType"`
	SentAt      time.Time `bson:"sentAt" json:"sentAt"`
	PhoneNumber string    `bson:"phoneNumber" json:"phoneNumber"`
}
