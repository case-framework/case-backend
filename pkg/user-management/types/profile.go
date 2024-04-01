package participantuser

import "go.mongodb.org/mongo-driver/bson/primitive"

type Profile struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Alias              string             `bson:"alias" json:"alias"`
	ConsentConfirmedAt int64              `bson:"consentConfirmedAt" json:"consentConfirmedAt"`
	CreatedAt          int64              `bson:"createdAt" json:"createdAt"`
	AvatarID           string             `bson:"avatarID" json:"avatarID"`
	MainProfile        bool               `bson:"mainProfile" json:"mainProfile"`
}
