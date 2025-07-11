package types

import "time"

type RenewToken struct {
	UserID     string    `bson:"userID,omitempty"`
	SessionID  string    `bson:"sessionID,omitempty"`
	RenewToken string    `bson:"renewToken,omitempty"`
	ExpiresAt  time.Time `bson:"expiresAt,omitempty"`
	NextToken  string    `bson:"nextToken,omitempty"` // token that replaces the current renew token
}
