package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type StudyMessage struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Type          string             `bson:"type,omitempty" json:"type,omitempty"`
	Payload       map[string]string  `bson:"payload,omitempty" json:"payload,omitempty"`
	ParticipantID string             `bson:"participantID,omitempty" json:"participantID,omitempty"`
}
