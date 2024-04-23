package types

type Account struct {
	Type               string           `bson:"type" json:"type"`
	AccountID          string           `bson:"accountID" json:"accountID"`
	AccountConfirmedAt int64            `bson:"accountConfirmedAt" json:"accountConfirmedAt"`
	Password           string           `bson:"password" json:"password"`
	AuthType           string           `bson:"authType" json:"authType"`
	VerificationCode   VerificationCode `bson:"verificationCode" json:"verificationCode"`
	PreferredLanguage  string           `bson:"preferredLanguage" json:"preferredLanguage"`

	// Rate limiting
	FailedLoginAttempts   []int64 `bson:"failedLoginAttempts" json:"failedLoginAttempts"`
	PasswordResetTriggers []int64 `bson:"passwordResetTriggers" json:"passwordResetTriggers"`
}

type VerificationCode struct {
	Code      string `bson:"code" json:"code"`
	Attempts  int64  `bson:"attempts" json:"attempts"`
	CreatedAt int64  `bson:"createdAt" json:"createdAt"`
	ExpiresAt int64  `bson:"expiresAt" json:"expiresAt"`
}
