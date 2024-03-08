package study

import "go.mongodb.org/mongo-driver/bson/primitive"

type SurveyResponse struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Key           string               `bson:"key" json:"key"`
	ParticipantID string               `bson:"participantID" json:"participantId"`
	VersionID     string               `bson:"versionID" json:"versionId"`
	OpenedAt      int64                `bson:"openedAt" json:"openedAt"`
	SubmittedAt   int64                `bson:"submittedAt" json:"submittedAt"`
	ArrivedAt     int64                `bson:"arrivedAt" json:"arrivedAt"`
	Responses     []SurveyItemResponse `bson:"responses" json:"responses"`
	Context       map[string]string    `bson:"context" json:"context"`
}

type SurveyItemResponse struct {
	Key  string       `bson:"key" json:"key"`
	Meta ResponseMeta `bson:"meta" json:"meta"`

	// for groups:
	Items []SurveyItemResponse `bson:"items,omitempty" json:"items,omitempty"`

	// for single items:
	Response         *ResponseItem `bson:"response,omitempty" json:"response,omitempty"`
	ConfidentialMode string        `bson:"confidentialMode,omitempty" json:"confidentialMode,omitempty"`
}

type ResponseMeta struct {
	Position   int32  `bson:"position" json:"position"`
	LocaleCode string `bson:"localeCode" json:"localeCode"`
	// timestamps
	Rendered  []int64 `bson:"rendered" json:"rendered"`
	Displayed []int64 `bson:"displayed" json:"displayed"`
	Responded []int64 `bson:"responded" json:"responded"`
}

type ResponseItem struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value,omitempty" json:"value,omitempty"`
	Dtype string `bson:"dtype,omitempty" json:"dtype,omitempty"`
	// for response option groups
	Items []*ResponseItem `bson:"items,omitempty" json:"items,omitempty"`
}
