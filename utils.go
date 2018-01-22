package goloquent

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// IsPrimaryKey :
func IsPrimaryKey(tag string) bool {
	isMatch, _ := regexp.MatchString(fmt.Sprintf("%s", tagKey), tag)
	return isMatch
}

// isNameReserved :
func isNameReserved(name string) bool {
	name = strings.TrimSpace(name)
	exp := strings.Replace(strings.Join(fieldNameReserved, "|"), "/", "\\/", -1)
	re := regexp.MustCompile(fmt.Sprintf("(?i)(%s)", exp))
	return re.MatchString(name)
}

func isEscape(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return false
	case reflect.Float32, reflect.Float64:
		return false
	case reflect.Bool:
		return false
	default:
		return true
	}
}

// isNumber :
func isNumber(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// IsNumeric :
// func IsNumeric(t reflect.Type) bool {
// 	switch t.Kind() {
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		return true
// 	case reflect.Float32, reflect.Float64:
// 		return true
// 	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		return true
// 	default:
// 		return false
// 	}
// }

// IsLegitName :
func isLegitName(name string) bool {
	name = strings.TrimSpace(name)
	re := regexp.MustCompile("(?i)^[a-z]+")
	return re.MatchString(name)
}
