package filter

import (
	"fmt"
	"strings"
	"time"
)

func queryFilter(operators []string, field *Field, list map[string]interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)
	opList := make(map[string]bool, 0)
	for _, item := range operators {
		opList[item] = true
	}

	for k, item := range list {
		k = strings.TrimSpace(strings.ToLower(k))
		if _, isExist := opList[k]; !isExist {
			return nil, fmt.Errorf(
				"filterable : invalid operator %q for %q", k, field.JSONName)
		}
		filters = append(
			filters,
			newFilter(field.Name, getOperator(k), item))
	}

	return filters, nil
}

func stringFilter(field *Field, it interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := it.(type) {
	case string:
		filters = append(
			filters,
			newFilter(field.Name, getOperator(OperatorEqual), val))

	case map[string]interface{}:
		f, err := queryFilter([]string{
			OperatorEqual,
			OperatorNotEqual,
			OperatorLike,
			OperatorNotLike,
			OperatorIn,
			OperatorNotIn,
		}, field, val)

		if err != nil {
			return nil, err
		}

		filters = append(filters, f...)

	default:
		return nil, fmt.Errorf(
			"filterable : invalid data type %T for int filter", val)
	}

	return filters, nil
}

func boolFilter(field *Field, it interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := it.(type) {
	case bool:
		filters = append(
			filters,
			newFilter(field.Name, getOperator(OperatorEqual), val))

	case map[string]interface{}:
		for k, item := range val {
			v, isOK := item.(bool)
			if !isOK {
				return nil, fmt.Errorf(
					"filterable : invalid data type %T int filter", item)
			}
			val[k] = v
		}

		f, err := queryFilter([]string{
			OperatorEqual,
			OperatorNotEqual,
		}, field, val)

		if err != nil {
			return nil, err
		}

		filters = append(filters, f...)

	default:
		return nil, fmt.Errorf(
			"filterable : invalid data type %T for int filter", val)
	}

	return filters, nil
}

func intFilter(field *Field, it interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := it.(type) {
	case float64:
		filters = append(
			filters,
			newFilter(field.Name, getOperator(OperatorEqual), int(val)))

	case map[string]interface{}:
		for k, item := range val {
			v, isOK := item.(float64)
			if !isOK {
				return nil, fmt.Errorf(
					"filterable : invalid data type %T int filter", item)
			}
			val[k] = int(v)
		}

		f, err := queryFilter([]string{
			OperatorEqual,
			OperatorNotEqual,
			OperatorGreaterThan,
			OperatorGreaterThanOrEqual,
			OperatorLessThan,
			OperatorLessThanOrEqual,
		}, field, val)

		if err != nil {
			return nil, err
		}

		filters = append(filters, f...)

	default:
		return nil, fmt.Errorf(
			"filterable : invalid data type %T for int filter", val)
	}

	return filters, nil
}

func floatFilter(field *Field, it interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := it.(type) {
	case float64:
		filters = append(
			filters,
			newFilter(field.Name, getOperator(OperatorEqual), val))

	case map[string]interface{}:
		for k, item := range val {
			v, isOK := item.(float64)
			if !isOK {
				return nil, fmt.Errorf(
					"filterable : invalid data type %T int filter", item)
			}
			val[k] = v
		}

		f, err := queryFilter([]string{
			OperatorEqual,
			OperatorNotEqual,
			OperatorGreaterThan,
			OperatorGreaterThanOrEqual,
			OperatorLessThan,
			OperatorLessThanOrEqual,
		}, field, val)

		if err != nil {
			return nil, err
		}

		filters = append(filters, f...)

	default:
		return nil, fmt.Errorf(
			"filterable : invalid data type %T for float filter", val)
	}

	return filters, nil
}

func timeFilter(field *Field, it interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := it.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		filters = append(filters, newFilter(field.Name, OperatorEqual, t))

	case map[string]interface{}:
		for k, item := range val {
			v, isOK := item.(string)
			if !isOK {
				return nil, fmt.Errorf(
					"filterable : invalid data type %T", item)
			}
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, err
			}
			val[k] = t
		}

		f, err := queryFilter([]string{
			OperatorEqual,
			OperatorNotEqual,
			OperatorGreaterThan,
			OperatorGreaterThanOrEqual,
			OperatorLessThan,
			OperatorLessThanOrEqual,
			OperatorIn,
			OperatorNotIn,
		}, field, val)

		if err != nil {
			return nil, err
		}

		filters = append(filters, f...)

	default:
		return nil, fmt.Errorf(
			"filterable : invalid data type %v for time filter", val)
	}

	return filters, nil
}
