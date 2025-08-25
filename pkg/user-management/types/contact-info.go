package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type ContactInfo struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type                   ContactInfoType    `bson:"type" json:"type"`
	ConfirmedAt            int64              `bson:"confirmedAt" json:"confirmedAt"`
	ConfirmationLinkSentAt int64              `bson:"confirmationLinkSentAt" json:"confirmationLinkSentAt"`
	Email                  string             `bson:"email,omitempty" json:"email,omitempty"`
	Phone                  string             `bson:"phone,omitempty" json:"phone,omitempty"`
}
