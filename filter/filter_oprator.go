package filter

// OperatorMappingList :
var OperatorMappingList = map[string]string{
	OperatorEqual:              "=",
	OperatorNotEqual:           "!=",
	OperatorGreaterThan:        ">",
	OperatorGreaterThanOrEqual: ">=",
	OperatorLessThan:           "<",
	OperatorLessThanOrEqual:    "<=",
	OperatorLike:               "like",
	OperatorNotLike:            "not like",
	OperatorIn:                 "in",
	OperatorNotIn:              "not in",
}

func getOperator(operator string) string {
	if o, isValid := OperatorMappingList[operator]; isValid {
		return o
	}

	return ""
}
