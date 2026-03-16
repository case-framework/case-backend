package types

import "go.mongodb.org/mongo-driver/v2/bson"

type ContactInfo struct {
	ID                     bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Type                   ContactInfoType `bson:"type" json:"type"`
	ConfirmedAt            int64           `bson:"confirmedAt" json:"confirmedAt"`
	ConfirmationLinkSentAt int64           `bson:"confirmationLinkSentAt" json:"confirmationLinkSentAt"`
	Email                  string          `bson:"email,omitempty" json:"email,omitempty"`
	Phone                  string          `bson:"phone,omitempty" json:"phone,omitempty"`
}
