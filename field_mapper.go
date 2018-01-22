package goloquent

import "reflect"

// Mapper :
type Mapper struct {
	StringFunc func(interface{}) (*string, error)
	ParseFunc  func(*Field, string) (interface{}, error)
}

var mappingList = map[reflect.Kind]*Mapper{
	reflect.String:  &Mapper{strToString, stringToStr},
	reflect.Bool:    &Mapper{boolToString, stringToBool},
	reflect.Int:     &Mapper{intToString, stringToInt},
	reflect.Int8:    &Mapper{intToString, stringToInt8},
	reflect.Int16:   &Mapper{intToString, stringToInt16},
	reflect.Int32:   &Mapper{intToString, stringToInt32},
	reflect.Int64:   &Mapper{int64ToString, stringToInt64},
	reflect.Float32: &Mapper{float32ToString, stringToFloat32},
	reflect.Float64: &Mapper{float64ToString, stringToFloat64},
	reflect.Slice:   &Mapper{sliceToString, stringToSlice},
	reflect.Array:   &Mapper{sliceToString, stringToSlice},
	reflect.Struct:  &Mapper{structToString, stringToStruct},
}
