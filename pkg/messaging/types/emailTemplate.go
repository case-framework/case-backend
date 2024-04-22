package types

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	EMAIL_TYPE_REGISTRATION                     = "registration"
	EMAIL_TYPE_INVITATION                       = "invitation"
	EMAIL_TYPE_VERIFY_EMAIL                     = "verify-email"
	EMAIL_TYPE_AUTH_VERIFICATION_CODE           = "verification-code"
	EMAIL_TYPE_PASSWORD_RESET                   = "password-reset"
	EMAIL_TYPE_PASSWORD_CHANGED                 = "password-changed"
	EMAIL_TYPE_ACCOUNT_ID_CHANGED               = "account-id-changed"
	EMAIL_TYPE_WEEKLY                           = "weekly"
	EMAIL_TYPE_STUDY_REMINDER                   = "study-reminder"
	EMAIL_TYPE_NEWSLETTER                       = "newsletter"
	EMAIL_TYPE_ACCOUNT_DELETED                  = "account-deleted"
	EMAIL_TYPE_ACCOUNT_DELETED_AFTER_INACTIVITY = "account-deleted-after-inactivity"
	EMAIL_TYPE_ACCOUNT_INACTIVITY               = "account-inactivity"
)

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
