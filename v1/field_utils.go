package v1

import (
	"fmt"
	"reflect"
)

func validateFields(t reflect.Type) error {

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fmt.Println(f)
	}

	return nil
}
