package study

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	SURVEY_ITEM_TYPE_PAGE_BREAK = "pageBreak"
	SURVEY_ITEM_TYPE_END        = "surveyEnd"
)

type SurveyItem struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key       string             `bson:"key" json:"key"`
	Follows   []string           `bson:"follows,omitempty" json:"follows,omitempty"`
	Condition *Expression        `bson:"condition,omitempty" json:"condition,omitempty"`
	Priority  float32            `bson:"priority,omitempty" json:"priority,omitempty"`

	Metadata map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`

	// Question group attributes
	Items           []SurveyItem `bson:"items,omitempty" json:"items,omitempty"`
	SelectionMethod *Expression  `bson:"selectionMethod,omitempty" json:"selectionMethod,omitempty"`

	// Question attributes
	Type             string         `bson:"type,omitempty" json:"type,omitempty"` // Specify some special types e.g. 'pageBreak','surveyEnd'
	Components       *ItemComponent `bson:"components,omitempty" json:"components,omitempty"`
	Validations      []Validation   `bson:"validations,omitempty" json:"validations,omitempty"`
	ConfidentialMode string         `bson:"confidentialMode,omitempty" json:"confidentialMode,omitempty"`
}

type Validation struct {
	Key  string     `bson:"key" json:"key"`
	Type string     `bson:"type" json:"type"` // kind of validation : 'soft' or 'hard'
	Rule Expression `bson:"expression" json:"rule"`
}

type ItemComponent struct {
	Role             string            `bson:"role" json:"role"`
	Key              string            `bson:"key" json:"key"`
	Content          []LocalisedObject `bson:"content" json:"content"`
	DisplayCondition *Expression       `bson:"displayCondition,omitempty" json:"displayCondition,omitempty"`
	Disabled         *Expression       `bson:"disabled,omitempty" json:"disabled,omitempty"`

	// group component
	Items []ItemComponent `bson:"items,omitempty" json:"items,omitempty"`
	Order *Expression     `bson:"order,omitempty" json:"order,omitempty"`

	// response compontent
	Dtype      string               `bson:"dtype,omitempty" json:"dtype,omitempty"`
	Properties *ComponentProperties `bson:"properties,omitempty" json:"properties,omitempty"`

	Style       []Style           `bson:"style,omitempty" json:"style,omitempty"`
	Description []LocalisedObject `bson:"description" json:"description"`
}

type Style struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

type ComponentProperties struct {
	Min           *ExpressionArg `bson:"min" json:"min"`
	Max           *ExpressionArg `bson:"max" json:"max"`
	StepSize      *ExpressionArg `bson:"stepSize" json:"stepSize"`
	DateInputMode *ExpressionArg `bson:"dateInputMode" json:"dateInputMode"`
}
