package filter

import (
	"reflect"
	"strings"
)

// Tag :
type Tag struct {
	Name     string
	JSONName string
}

func newTag(sf reflect.StructField) *Tag {
	jsonName := sf.Name
	strTag := strings.TrimSpace(sf.Tag.Get(tagJSON))
	if strTag != "" {
		tagPath := strings.Split(strTag, ",")
		jsonName = strings.TrimSpace(tagPath[0])
	}

	name := sf.Name
	strTag = strings.TrimSpace(sf.Tag.Get(tagFilter))
	if strTag != "" {
		tagPath := strings.Split(strTag, ",")
		name = strings.TrimSpace(tagPath[0])
	}

	return &Tag{
		JSONName: jsonName,
		Name:     name,
	}
}
