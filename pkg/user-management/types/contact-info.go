package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type ContactInfo struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type                   string             `bson:"type" json:"type"`
	ConfirmedAt            int64              `bson:"confirmedAt" json:"confirmedAt"`
	ConfirmationLinkSentAt int64              `bson:"confirmationLinkSentAt" json:"confirmationLinkSentAt"`
	Email                  string             `bson:"email" json:"email"`
	Phone                  string             `bson:"phone" json:"phone"`
}
