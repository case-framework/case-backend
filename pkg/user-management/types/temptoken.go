package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TOKEN_PURPOSE_INVITATION                 = "invitation"
	TOKEN_PURPOSE_PASSWORD_RESET             = "password-reset"
	TOKEN_PURPOSE_CONTACT_VERIFICATION       = "contact-verification"
	TOKEN_PURPOSE_SURVEY_LOGIN               = "survey-login"
	TOKEN_PURPOSE_UNSUBSCRIBE_NEWSLETTER     = "unsubscribe-newsletter"
	TOKEN_PURPOSE_RESTORE_ACCOUNT_ID         = "restore_account_id"
	TOKEN_PURPOSE_INACTIVE_USER_NOTIFICATION = "inactive-user-notification"
)

type TempToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"token_id,omitempty"`
	Token      string             `bson:"token" json:"token"`
	Expiration time.Time          `bson:"expiration" json:"expiration"`
	Purpose    string             `bson:"purpose" json:"purpose"`
	UserID     string             `bson:"userID" json:"userID"`
	Info       map[string]string  `bson:"info" json:"info"`
	InstanceID string             `bson:"instanceID" json:"instanceID"`
}
