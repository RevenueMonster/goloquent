package goloquent

import (
	"fmt"
	"reflect"
)

type filterFunc func(*FilterField, interface{}) ([]Filter, error)

var typeFilterList = map[reflect.Type]filterFunc{
	typeOfString:       stringFilter,
	typeOfBool:         boolFilter,
	typeOfInt:          intFilter,
	typeOfInt8:         intFilter,
	typeOfInt16:        intFilter,
	typeOfInt32:        intFilter,
	typeOfInt64:        int64Filter,
	typeOfFloat32:      float32Filter,
	typeOfFloat64:      float64Filter,
	typeOfTime:         timeFilter,
	typeOfDataStoreKey: keyFilter,
}

func isValidType(t reflect.Type) (filterFunc, error) {
	var parseFunc filterFunc

	switch t.Kind() {
	case reflect.String:
		goto routineValidType

	case reflect.Bool:
		goto routineValidType

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		goto routineValidType

	case reflect.Float32, reflect.Float64:
		goto routineValidType

	case reflect.Struct:
		goto routineValidType

	case reflect.Ptr:
		t = t.Elem()
		if t == typeOfDataStoreKey {
			goto routineValidType
		}
		fallthrough

	default:
		return nil, fmt.Errorf("goloquent: invalid data type %q in filter layout", t.Kind())
	}

routineValidType:
	parseFunc, isExist := typeFilterList[t]
	if !isExist {
		return nil, nil
	}

	return parseFunc, nil
}
