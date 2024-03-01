package messaging

import "go.mongodb.org/mongo-driver/bson/primitive"

// email template
type EmailTemplate struct {
	ID              primitive.ObjectID  `bson:"_id" json:"id,omitempty"`
	MessageType     string              `bson:"messageType" json:"messageType"`
	StudyKey        string              `bson:"studyKey,omitempty" json:"studyKey"`
	DefaultLanguage string              `bson:"defaultLanguage" json:"defaultLanguage"`
	HeaderOverrides *HeaderOverrides    `bson:"headerOverrides" json:"headerOverrides"`
	Translations    []LocalizedTemplate `bson:"translations" json:"translations"`
}

type HeaderOverrides struct {
	From      string   `bson:"from" json:"from"`
	Sender    string   `bson:"sender" json:"sender"`
	ReplyTo   []string `bson:"replyTo" json:"replyTo"`
	NoReplyTo bool     `bson:"noReplyTo" json:"noReplyTo"`
}

type LocalizedTemplate struct {
	Lang        string `bson:"languageCode" json:"lang"`
	Subject     string `bson:"subject" json:"subject"`
	TemplateDef string `bson:"templateDef" json:"templateDef"`
}

// messageschedule

// outgoing email

// sent email
