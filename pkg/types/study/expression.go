package study

type Expression struct {
	Name       string          `bson:"name" json:"name"` // Name of the operation to be evaluated
	ReturnType string          `bson:"returnType,omitempty" json:"returnType"`
	Data       []ExpressionArg `bson:"data,omitempty" json:"data"` // Operation arguments
}

type ExpressionArg struct {
	DType string      `bson:"dtype" json:"dtype"`
	Exp   *Expression `bson:"exp,omitempty" json:"exp,omitempty"`
	Str   string      `bson:"str,omitempty" json:"str,omitempty"`
	Num   float64     `bson:"num,omitempty" json:"num,omitempty"`
}
