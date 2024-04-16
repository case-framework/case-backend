package types

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	PARTICIPANT_STUDY_STATUS_ACTIVE          = "active"
	PARTICIPANT_STUDY_STATUS_TEMPORARY       = "temporary" // for participants without a registered account
	PARTICIPANT_STUDY_STATUS_EXITED          = "exited"
	PARTICIPANT_STUDY_STATUS_ACCOUNT_DELETED = "accountDeleted"
)

// Participant defines the datamodel for current state of the participant in a study as stored in the database
type Participant struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	ParticipantID       string               `bson:"participantID" json:"participantId"` // reference to the study specific participant ID
	CurrentStudySession string               `bson:"currentStudySession" json:"currentStudySession"`
	EnteredAt           int64                `bson:"enteredAt" json:"enteredAt"`
	StudyStatus         string               `bson:"studyStatus" json:"studyStatus"` // shows if participant is active in the study - possible values: "active", "temporary", "exited". Other values are possible and are handled like "exited" on the server.
	Flags               map[string]string    `bson:"flags" json:"flags"`
	AssignedSurveys     []AssignedSurvey     `bson:"assignedSurveys" json:"assignedSurveys"`
	LastSubmissions     map[string]int64     `bson:"lastSubmission" json:"lastSubmissions"` // surveyKey with timestamp
	Messages            []ParticipantMessage `bson:"messages" json:"messages"`
}

type ParticipantMessage struct {
	ID           string `bson:"id" json:"id"`
	Type         string `bson:"type" json:"type"`
	ScheduledFor int64  `bson:"scheduledFor" json:"scheduledFor"`
}
