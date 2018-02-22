package filter

import (
	"fmt"
	"reflect"
	"time"
)

type filterFunc func(*Field, interface{}) ([]Filter, error)

// Filter :
type Filter struct {
	Name     string
	Operator string
	Value    interface{}
}

func newFilter(name, operator string, value interface{}) Filter {
	return Filter{
		Name:     name,
		Operator: operator,
		Value:    value,
	}
}

var (
	typeOfString  = reflect.TypeOf(string(""))
	typeOfBool    = reflect.TypeOf(bool(false))
	typeOfInt     = reflect.TypeOf(int(0))
	typeOfInt8    = reflect.TypeOf(int8(0))
	typeOfInt16   = reflect.TypeOf(int16(0))
	typeOfInt32   = reflect.TypeOf(int32(0))
	typeOfInt64   = reflect.TypeOf(int64(0))
	typeOfUint    = reflect.TypeOf(uint(0))
	typeOfUint8   = reflect.TypeOf(uint8(0))
	typeOfUint16  = reflect.TypeOf(uint16(0))
	typeOfUint32  = reflect.TypeOf(uint32(0))
	typeOfUint64  = reflect.TypeOf(uint64(0))
	typeOfFloat32 = reflect.TypeOf(float32(0))
	typeOfFloat64 = reflect.TypeOf(float64(0))
	typeOfTime    = reflect.TypeOf(time.Time{})
)

var typeFilterList = map[reflect.Type]filterFunc{
	typeOfString:  stringFilter,
	typeOfBool:    boolFilter,
	typeOfInt:     intFilter,
	typeOfInt8:    intFilter,
	typeOfInt16:   intFilter,
	typeOfInt32:   intFilter,
	typeOfInt64:   intFilter,
	typeOfUint:    intFilter,
	typeOfUint8:   intFilter,
	typeOfUint16:  intFilter,
	typeOfUint32:  intFilter,
	typeOfUint64:  intFilter,
	typeOfFloat32: floatFilter,
	typeOfFloat64: floatFilter,
	typeOfTime:    timeFilter,
}

func isValidType(t reflect.Type) (filterFunc, error) {
	switch t.Kind() {
	case reflect.String:
		goto routineValidType

	case reflect.Bool:
		goto routineValidType

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		goto routineValidType

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		goto routineValidType

	case reflect.Float32, reflect.Float64:
		goto routineValidType

	case reflect.Struct:
		goto routineValidType

	default:
		return nil, fmt.Errorf("filterable : invalid data type %T", t)
	}

routineValidType:
	mapFunc, isValid := typeFilterList[t]
	if !isValid {
		return mapFunc, nil
	}

	return mapFunc, nil
}
