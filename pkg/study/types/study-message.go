package types

import "go.mongodb.org/mongo-driver/v2/bson"

type StudyMessage struct {
	ID            bson.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	Type          string            `bson:"type,omitempty" json:"type,omitempty"`
	Payload       map[string]string `bson:"payload,omitempty" json:"payload,omitempty"`
	ParticipantID string            `bson:"participantID,omitempty" json:"participantID,omitempty"`
}
