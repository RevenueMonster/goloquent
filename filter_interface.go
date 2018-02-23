package goloquent

import (
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

func queryFilter(operators []string, field *FilterField, list map[string]interface{}) ([]Filter, error) {
	opList := make(map[string]bool, 0)
	for _, item := range operators {
		opList[item] = true
	}

	filters := make([]Filter, 0)
	for k, item := range list {
		k = strings.TrimSpace(strings.ToLower(k))
		if _, isExist := opList[k]; !isExist {
			return nil, fmt.Errorf(
				"goloquent: invalid operator %q for %q", k, field.JSONName)
		}
		f := newFilter(field.Name, getOperator(k), item)
		filters = append(filters, f)
	}

	return filters, nil
}

func stringFilter(field *FilterField, it interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := it.(type) {
	case string:
		f := newFilter(field.Name, getOperator(OperatorEqual), val)
		filters = append(filters, f)

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
			"goloquent: invalid data type %T for string filter", val)
	}

	return filters, nil
}

func boolFilter(field *FilterField, it interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := it.(type) {
	case bool:
		f := newFilter(field.Name, getOperator(OperatorEqual), val)
		filters = append(filters, f)

	case map[string]interface{}:
		for k, item := range val {
			v, isOK := item.(bool)
			if !isOK {
				return nil, fmt.Errorf(
					"goloquent: invalid data type %T bool filter", item)
			}
			val[k] = v
		}

	default:
		return nil, fmt.Errorf(
			"goloquent: invalid data type %T for bool filter", val)
	}

	return filters, nil
}

func intFilter(field *FilterField, value interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := value.(type) {
	case float64:
		f := newFilter(field.Name, getOperator(OperatorEqual), int(val))
		filters = append(filters, f)

	case map[string]interface{}:

	default:
		return nil, fmt.Errorf(
			"goloquent: invalid data type %v for int filter", val)
	}

	return filters, nil
}

func int64Filter(field *FilterField, value interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := value.(type) {
	case float64:
		f := newFilter(field.Name, getOperator(OperatorEqual), int64(val))
		filters = append(filters, f)

	case map[string]interface{}:

	default:
		return nil, fmt.Errorf(
			"goloquent: invalid data type %v for int64 filter", val)
	}

	return filters, nil
}

func float32Filter(field *FilterField, value interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := value.(type) {
	case float64:
		f := newFilter(field.Name, getOperator(OperatorEqual), float32(val))
		filters = append(filters, f)

	case map[string]interface{}:

	default:
		return nil, fmt.Errorf(
			"goloquent: invalid data type %v for float32 filter", val)
	}

	return filters, nil
}

func float64Filter(field *FilterField, value interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := value.(type) {
	case float64:
		f := newFilter(field.Name, getOperator(OperatorEqual), val)
		filters = append(filters, f)

	case map[string]interface{}:

	default:
		return nil, fmt.Errorf(
			"goloquent: invalid data type %v for float64 filter", val)
	}

	return filters, nil
}

func timeFilter(field *FilterField, value interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := value.(type) {
	case string:
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		f := newFilter(field.Name, getOperator(OperatorEqual), t)
		filters = append(filters, f)

	case map[string]interface{}:
		for k, item := range val {
			v, isOK := item.(string)
			if !isOK {
				return nil, fmt.Errorf(
					"goloquent: invalid data type %T", item)
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
			"goloquent: invalid data type %v for time filter", val)
	}

	return filters, nil
}

func keyFilter(field *FilterField, value interface{}) ([]Filter, error) {
	filters := make([]Filter, 0)

	switch val := value.(type) {
	case nil:
		f := newFilter(field.Name, getOperator(OperatorEqual), nil)
		filters = append(filters, f)

	case string:
		key, err := datastore.DecodeKey(val)
		if err != nil {
			return nil, err
		}
		f := newFilter(field.Name, getOperator(OperatorEqual), key)
		filters = append(filters, f)

	case map[string]interface{}:

	case []interface{}:
		keys := make([]*datastore.Key, 0)
		for _, item := range val {
			strKey, isOK := item.(string)
			if !isOK {
				return nil, fmt.Errorf("goloquent: invalid data type %T", item)
			}
			k, err := datastore.DecodeKey(strKey)
			if err != nil {
				return nil, err
			}
			keys = append(keys, k)
		}

		f := newFilter(field.Name, getOperator(OperatorIn), keys)
		filters = append(filters, f)

	default:
		return nil, fmt.Errorf(
			"goloquent: invalid data type %T for key filter", val)
	}

	return filters, nil
}
