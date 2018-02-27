package goloquent

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// FilterField :
type FilterField struct {
	Name     string
	JSONName string
	Type     reflect.Type
	Index    []int
	mapFunc  filterFunc
}

// Map :
func (f *FilterField) Map(it interface{}) ([]Filter, error) {
	return f.mapFunc(f, it)
}

func newFilterField(name, jsonName string, mapFunc filterFunc) *FilterField {
	return &FilterField{
		Name:     name,
		JSONName: jsonName,
		mapFunc:  mapFunc,
	}
}

// ParseFilter :
func ParseFilter(layout interface{}, input []byte) ([]Filter, error) {

	if reflect.TypeOf(layout).Kind() != reflect.Struct {
		return nil, errors.New("goloquent: filter layout must be struct")
	}

	fieldList, err := listFilter(layout)
	if err != nil {
		return nil, err
	}

	jsonByte := input
	str, err := strconv.Unquote(string(input))
	if err == nil {
		jsonByte = []byte(str)
	}

	m := make(map[string]interface{}, 0)
	if err := json.Unmarshal(jsonByte, &m); err != nil {
		return nil, err
	}

	filters := make([]Filter, 0)
	for k, item := range m {
		field, isExist := fieldList[k]
		if !isExist {
			return nil, fmt.Errorf("goloquent: filter value has invalid json key %q", k)
		}
		f, err := field.Map(item)
		if err != nil {
			return nil, err
		}

		filters = append(filters, f...)
	}

	return filters, nil
}
