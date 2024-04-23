package types

import "time"

type OTP struct {
	Code      string    `bson:"code" json:"code"`
	UserID    string    `bson:"userID" json:"userID"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	Type      string    `bson:"type" json:"type"`
}
