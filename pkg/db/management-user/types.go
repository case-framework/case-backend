package managementuser

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Permission struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	SubjectID    string             `json:"subject_id,omitempty" bson:"subjectId,omitempty"`
	SubjectType  string             `json:"subject_type,omitempty" bson:"subjectType,omitempty"`
	ResourceType string             `json:"resource_type,omitempty" bson:"resourceType,omitempty"`
	ResourceKey  string             `json:"resource_key,omitempty" bson:"resourceKey,omitempty"`
	Action       string             `json:"action,omitempty" bson:"action,omitempty"`
	Limiter      string             `json:"limiter,omitempty" bson:"limiter,omitempty"`
}

// SubjectType is the type of the subject e.g., user or service
// ResourceType is the type of the resource e.g., messages, studies, etc.
// ResourceKey is the key of the resource e.g., the study id, or * for all
// Limiter is an optional additional criteria for the permission e.g., survey keys, or message types
// Action is the action that is allowed e.g., download_responses, upload_survey, etc.

type ManagementUser struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Sub         string             `json:"sub,omitempty" bson:"sub,omitempty"`
	Email       string             `json:"email,omitempty" bson:"email,omitempty"`
	Username    string             `json:"username,omitempty" bson:"username,omitempty"`
	LastLoginAt time.Time          `json:"lastLoginAt,omitempty" bson:"lastLoginAt,omitempty"`
	CreatedAt   time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type Session struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	RenewToken string             `json:"renewToken,omitempty" bson:"renewToken,omitempty"`
	CreatedAt  time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}
