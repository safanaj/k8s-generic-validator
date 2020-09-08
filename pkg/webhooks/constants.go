package webhooks

type Operator = string

const (
	// for string and equality in genral (==)
	OperatorIs    Operator = "Is"
	OperatorIsNot Operator = "IsNot"
	// for string
	OperatorIn    Operator = "In"
	OperatorNotIn Operator = "NotIn"
	// for numeric
	// >
	OperatorGreaterThan Operator = "GreaterThan"
	OperatorMoreThan    Operator = "MoreThan"
	// <
	OperatorSmallerThan Operator = "SmallerThan"
	OperatorLessThan    Operator = "LessThan"
	// >=
	OperatorEqualOrGreaterThan Operator = "EqualOrGreaterThan"
	OperatorEqualOrMoreThan    Operator = "EqualOrMoreThan"
	// <=
	OperatorEqualOrSmallerThan Operator = "EqualOrSmallerThan"
	OperatorEqualOrLessThan    Operator = "EqualOrLessThan"
)

type ValueType = string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeBool    ValueType = "bool"
	ValueTypeInt     ValueType = "int"
	ValueTypeInt64   ValueType = "int64"
	ValueTypeFloat   ValueType = "float"
	ValueTypeFloat64 ValueType = "float64"

	// ValueTypeStringSlice  ValueType = "[]string"
	// ValueTypeBoolSlice    ValueType = "[]bool"
	// ValueTypeIntSlice     ValueType = "[]int"
	// ValueTypeInt64Slice   ValueType = "[]int64"
	// ValueTypeFloatSlice   ValueType = "[]float"
	// ValueTypeFloat64Slice ValueType = "[]float64"
)
