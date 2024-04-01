package participantuser

import "time"

type RenewToken struct {
	UserID     string    `bson:"userID"`
	RenewToken string    `bson:"renewToken"`
	ExpiresAt  time.Time `bson:"expiresAt"`
	NextToken  string    `bson:"nextToken"` // token that replaces the current renew token
}
