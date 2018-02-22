package goloquent

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// ParseFilter :
func ParseFilter(layout interface{}, b []byte) ([]Filter, error) {

	if reflect.TypeOf(layout).Kind() != reflect.Struct {
		return nil, errors.New("goloquent: filter layout must be struct")
	}

	fieldList, err := getFilter(layout)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{}, 0)
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	// filters := make([]Filter, 0)
	for k, item := range m {
		field, isExist := fieldList[k]
		if !isExist {
			return nil, fmt.Errorf("goloquent : filter has invalid json key %q", k)
		}
		// f, err := field.Map(item)
		// if err != nil {
		// 	return nil, err
		// }
		// filters = append(filters, f...)
		fmt.Println(k, item, field)
	}

	return nil, nil
}
