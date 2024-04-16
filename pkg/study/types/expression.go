package study

type Expression struct {
	Name       string          `bson:"name" json:"name"` // Name of the operation to be evaluated
	ReturnType string          `bson:"returnType,omitempty" json:"returnType,omitempty"`
	Data       []ExpressionArg `bson:"data,omitempty" json:"data,omitempty"` // Operation arguments
}

type ExpressionArg struct {
	DType string      `bson:"dtype" json:"dtype"`
	Exp   *Expression `bson:"exp,omitempty" json:"exp,omitempty"`
	Str   string      `bson:"str,omitempty" json:"str,omitempty"`
	Num   float64     `bson:"num,omitempty" json:"num,omitempty"`
}

func (exp ExpressionArg) IsExpression() bool {
	return exp.DType == "exp"
}

func (exp ExpressionArg) IsNumber() bool {
	return exp.DType == "num"
}

func (exp ExpressionArg) IsString() bool {
	return exp.DType == "str"
}
