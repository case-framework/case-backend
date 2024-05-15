package types

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	SURVEY_AVAILABLE_FOR_PUBLIC                   = "public"
	SURVEY_AVAILABLE_FOR_TEMPORARY_PARTICIPANTS   = "temporary_participants"
	SURVEY_AVAILABLE_FOR_ACTIVE_PARTICIPANTS      = "active_participants"
	SURVEY_AVAILABLE_FOR_PARTICIPANTS_IF_ASSIGNED = "participants_if_assigned"
)

type Survey struct {
	ID                           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SurveyKey                    string             `bson:"surveyKey,omitempty" json:"surveyKey,omitempty"`
	Props                        SurveyProps        `bson:"props,omitempty" json:"props,omitempty"`
	PrefillRules                 []Expression       `bson:"prefillRules,omitempty" json:"prefillRules,omitempty"`
	ContextRules                 *SurveyContextDef  `bson:"contextRules,omitempty" json:"contextRules,omitempty"`
	MaxItemsPerPage              *MaxItemsPerPage   `bson:"maxItemsPerPage,omitempty" json:"maxItemsPerPage,omitempty"`
	AvailableFor                 string             `bson:"availableFor,omitempty" json:"availableFor,omitempty"`
	RequireLoginBeforeSubmission bool               `bson:"requireLoginBeforeSubmission,omitempty" json:"requireLoginBeforeSubmission,omitempty"`

	Published        int64             `bson:"published,omitempty" json:"published,omitempty"`
	Unpublished      int64             `bson:"unpublished,omitempty" json:"unpublished,omitempty"`
	SurveyDefinition SurveyItem        `bson:"surveyDefinition,omitempty" json:"surveyDefinition,omitempty"`
	VersionID        string            `bson:"versionID,omitempty" json:"versionId,omitempty"`
	Metadata         map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

type SurveyProps struct {
	Name            []LocalisedObject `bson:"name,omitempty" json:"name,omitempty"`
	Description     []LocalisedObject `bson:"description,omitempty" json:"description,omitempty"`
	TypicalDuration []LocalisedObject `bson:"typicalDuration,omitempty" json:"typicalDuration,omitempty"`
}

type MaxItemsPerPage struct {
	Large int32 `bson:"large,omitempty" json:"large,omitempty"`
	Small int32 `bson:"small,omitempty" json:"small,omitempty"`
}

type SurveyContextDef struct {
	Mode              *ExpressionArg `bson:"mode,omitempty" json:"mode,omitempty"`
	PreviousResponses []Expression   `bson:"previousResponses,omitempty" json:"previousResponses,omitempty"`
}
