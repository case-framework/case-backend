package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type OutgoingEmail struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MessageType     string             `bson:"messageType" json:"messageType"`
	To              []string           `bson:"to" json:"to"`
	Subject         string             `bson:"subject" json:"subject"`
	HeaderOverrides *HeaderOverrides   `bson:"headerOverrides" json:"headerOverrides"`
	Content         string             `bson:"content" json:"content"`
	AddedAt         int64              `bson:"addedAt" json:"addedAt"`
	HighPrio        bool               `bson:"highPrio" json:"highPrio"`
	LastSendAttempt int64              `bson:"lastSendAttempt" json:"lastSendAttempt"`
}
