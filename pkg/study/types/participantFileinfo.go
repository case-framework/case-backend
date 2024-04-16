package types

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	FILE_STATUS_UPLOADING = "uploading"
	FILE_STATUS_READY     = "ready"
)

type FileInfo struct {
	ID                   primitive.ObjectID    `bson:"_id,omitempty" json:"id,omitempty"`
	ParticipantID        string                `bson:"participantID,omitempty" json:"participantID,omitempty"`
	Status               string                `bson:"status,omitempty" json:"status,omitempty"`
	UploadedBy           string                `bson:"uploadedBy,omitempty" json:"uploadedBy,omitempty"` // if not uploaded by the participant
	Path                 string                `bson:"path,omitempty" json:"path,omitempty"`
	PreviewPath          string                `bson:"previewPath,omitempty" json:"previewPath,omitempty"`
	SubStudy             string                `bson:"subStudy,omitempty" json:"subStudy,omitempty"`
	SubmittedAt          int64                 `bson:"submittedAt,omitempty" json:"submittedAt,omitempty"`
	FileType             string                `bson:"fileType,omitempty" json:"fileType,omitempty"`
	VisibleToParticipant bool                  `bson:"visibleToParticipant,omitempty" json:"visibleToParticipant,omitempty"`
	Name                 string                `bson:"name,omitempty" json:"name,omitempty"`
	Size                 int32                 `bson:"size,omitempty" json:"size,omitempty"`
	ReferencedIn         []FileObjectReference `bson:"referencedIn,omitempty" json:"referencedIn,omitempty"`
}

type FileObjectReference struct {
	ID   string `bson:"id,omitempty" json:"id,omitempty"`
	Type string `bson:"type,omitempty" json:"type,omitempty"`
	Time int64  `bson:"time,omitempty" json:"time,omitempty"`
}

func (f *FileInfo) AddReference(ref FileObjectReference) error {
	f.ReferencedIn = append(f.ReferencedIn, ref)
	return nil
}

func (f *FileInfo) RemoveReference(refID string) error {
	for i, cf := range f.ReferencedIn {
		if cf.ID == refID {
			f.ReferencedIn = append(f.ReferencedIn[:i], f.ReferencedIn[i+1:]...)
			return nil
		}
	}
	return errors.New("reference not found")
}
