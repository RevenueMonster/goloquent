package goloquent

import (
	"reflect"
	"strings"
)

// FilterTag :
type FilterTag struct {
	Name     string
	JSONName string
}

func newFilterTag(sf reflect.StructField) *FilterTag {
	jsonName := sf.Name
	strTag := strings.TrimSpace(sf.Tag.Get("json"))
	if strTag != "" {
		tagPath := strings.Split(strTag, ",")
		jsonName = strings.TrimSpace(tagPath[0])
	}

	name := sf.Name
	strTag = strings.TrimSpace(sf.Tag.Get(optionTagGoloquent))
	if strTag != "" {
		tagPath := strings.Split(strTag, ",")
		name = strings.TrimSpace(tagPath[0])
	}

	return &FilterTag{
		JSONName: jsonName,
		Name:     name,
	}
}
