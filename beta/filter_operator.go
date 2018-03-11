package goloquent

var operatorList = map[string]string{
	OperatorEqual:              "=",
	OperatorNotEqual:           "!=",
	OperatorGreaterThan:        ">",
	OperatorGreaterThanOrEqual: ">=",
	OperatorLessThan:           "<",
	OperatorLessThanOrEqual:    "<=",
	OperatorLike:               "LIKE",
	OperatorNotLike:            "NOT LIKE",
	OperatorIn:                 "IN",
	OperatorNotIn:              "NOT IN",
}

func getOperator(o string) string {
	if op, isValid := operatorList[o]; isValid {
		return op
	}

	return ""
}
