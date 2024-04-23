package types

import "time"

type OTP struct {
	Code      string    `bson:"code" json:"code"`
	UserID    string    `bson:"userID" json:"userID"`
	ExpiresAt time.Time `bson:"expiresAt" json:"expiresAt"`
	Type      string    `bson:"type" json:"type"`
}
