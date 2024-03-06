package study

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StudyRules struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	StudyKey        string             `bson:"studyKey" json:"studyKey"`
	UploadedAt      int64              `bson:"uploadedAt" json:"uploadedAt"`
	UploadedBy      string             `bson:"uploadedBy" json:"uploadedBy"`
	Rules           []Expression       `bson:"rules,omitempty" json:"rules"`
	SerialisedRules string             `bson:"serialisedRules,omitempty" json:"serialisedRules,omitempty"`
}

func (studyRules *StudyRules) MarshalRules() error {
	rulesString, err := json.Marshal(studyRules.Rules)
	if err != nil {
		return err
	}
	studyRules.SerialisedRules = string(rulesString)
	studyRules.Rules = nil
	return nil
}

func (studyRules *StudyRules) UnmarshalRules() error {
	if studyRules.SerialisedRules == "" {
		if studyRules.Rules == nil {
			studyRules.Rules = []Expression{}
			return nil
		}
		return nil
	}
	var rules []Expression
	err := json.Unmarshal([]byte(studyRules.SerialisedRules), &rules)
	if err != nil {
		return err
	}
	studyRules.Rules = rules
	studyRules.SerialisedRules = ""
	return nil
}
