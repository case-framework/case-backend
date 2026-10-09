package types

type LocalisedObject struct {
	Code string `bson:"code" json:"code"`
	// For texts
	Parts []ExpressionArg `bson:"parts" json:"parts"`
}
