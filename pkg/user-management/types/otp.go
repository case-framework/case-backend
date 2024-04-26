package types

import "time"

type OTPType string

const (
	EmailOTP OTPType = "email"
	SMSOTP   OTPType = "sms"
)

type OTP struct {
	Code      string    `bson:"code" json:"code"`
	UserID    string    `bson:"userID" json:"userID"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	Type      OTPType   `bson:"type" json:"type"`
}
