package types

type AssignedSurvey struct {
	StudyKey   string `bson:"studyKey" json:"studyKey"`
	SurveyKey  string `bson:"surveyKey" json:"surveyKey"`
	ValidFrom  int64  `bson:"validFrom" json:"validFrom"`
	ValidUntil int64  `bson:"validUntil" json:"validUntil"`
	Category   string `bson:"category" json:"category"`
	ProfileID  string `bson:"profileID" json:"profileID"` // optional when sending surveys to multiple profiles
}
