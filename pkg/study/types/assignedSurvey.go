package types

const (
	ASSIGNED_SURVEY_CATEGORY_PRIO   = "prio"
	ASSIGNED_SURVEY_CATEGORY_NORMAL = "normal"
	ASSIGNED_SURVEY_CATEGORY_QUICK  = "quick"
	ASSIGNED_SURVEY_CATEGORY_UPDATE = "update"
)

type AssignedSurvey struct {
	SurveyKey  string `bson:"surveyKey" json:"surveyKey"`
	ValidFrom  int64  `bson:"validFrom" json:"validFrom"`
	ValidUntil int64  `bson:"validUntil" json:"validUntil"`
	Category   string `bson:"category" json:"category"`
	ProfileID  string `bson:"profileID" json:"profileID"` // optional when sending surveys to multiple profiles
}
