package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type operatorMapper struct {
	Compatible string
	StringFunc func(interface{}) (*string, error)
}

func anyOperatorToString(it interface{}) (*string, error) {
	t := reflect.TypeOf(it)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	stringFunc, isValid := interfaceToStringList[t]
	if !isValid {
		return nil, ErrUnsupportDataType
	}

	str, err := stringFunc(it)
	if err != nil {
		return nil, err
	}

	if str == nil {
		return nil, nil
	}

	if isEscape(t) {
		e := fmt.Sprintf("%q", *str)
		return &e, nil

	}

	return str, nil
}

func eqOperatorToString(it interface{}) (*string, error) {
	if it == nil {
		return nil, nil
	}
	return anyOperatorToString(it)
}

func compareOperatorToString(it interface{}) (*string, error) {
	var errNotAllowNull = errors.New("goloquent: nil value is not allow for comparison operators")
	if it == nil {
		return nil, errNotAllowNull
	}

	str, err := anyOperatorToString(it)
	if err != nil {
		return nil, err
	}

	if str == nil {
		return nil, errNotAllowNull
	}

	return str, nil
}

func inOperatorToString(it interface{}) (*string, error) {
	var (
		str *string
		err error
	)

	t := reflect.TypeOf(it)
	v := reflect.Indirect(reflect.ValueOf(it))

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		values := make(map[string]bool, 0)
		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)
			if f.Interface() == nil {
				values["NULL"] = true
				continue
			}

			str, err := anyOperatorToString(f.Interface())
			if err != nil {
				return nil, err
			}

			if str == nil {
				values["NULL"] = true
				continue
			}

			values[*str] = true
		}

		q := make([]string, 0)
		for k := range values {
			q = append(q, k)
		}
		st := fmt.Sprintf("(%s)", strings.Join(q, ","))
		str = &st

	default:
		err = errors.New("goloquent: invalid datatype, 'in' only support slice nor array")

	}

	if err != nil {
		return nil, err
	}

	return str, nil
}
