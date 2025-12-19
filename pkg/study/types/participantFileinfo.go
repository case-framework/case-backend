package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	FILE_STATUS_UPLOADING = "uploading"
	FILE_STATUS_READY     = "ready"
)

type FileInfo struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ParticipantID string             `bson:"participantID,omitempty" json:"participantID,omitempty"`
	Status        string             `bson:"status,omitempty" json:"status,omitempty"`
	UploadedBy    string             `bson:"uploadedBy,omitempty" json:"uploadedBy,omitempty"` // if not uploaded by the participant
	Path          string             `bson:"path,omitempty" json:"path,omitempty"`
	PreviewPath   string             `bson:"previewPath,omitempty" json:"previewPath,omitempty"`

	SubmittedAt int64     `bson:"submittedAt,omitempty" json:"submittedAt,omitempty"` // deprecated, use CreatedAt instead
	CreatedAt   time.Time `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt   time.Time `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`

	FileType             string `bson:"fileType,omitempty" json:"fileType,omitempty"`
	VisibleToParticipant bool   `bson:"visibleToParticipant" json:"visibleToParticipant"`
	Size                 int64  `bson:"size,omitempty" json:"size,omitempty"`
}
