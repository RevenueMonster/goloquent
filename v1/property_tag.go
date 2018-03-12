package v1

import (
	"reflect"
	"strings"
)

const (
	tagKey       = "__key__"
	tagOmitEmpty = "omitempty"
	tagNoIndex   = "noindex"
	tagFlatten   = "flatten"
	tagNullable  = "nullable"
	tagUnsigned  = "unsigned" // extra
	tagUnique    = "unique"   // extra
	tagLongText  = "longtext" // extra
)

// Tag :
type Tag struct {
	Name         string          // Tag name define in option tag
	FieldName    string          // Original name in struct field
	options      map[string]bool // Option tag define by user
	isPrimaryKey bool
}

// IsSkip : return true if option tag have "-" sign
func (t *Tag) IsSkip() bool {
	return t.Name == "-"
}

// IsPrimaryKey :
func (t *Tag) IsPrimaryKey() bool {
	return t.isPrimaryKey
}

// IsFlatten :
func (t *Tag) IsFlatten() bool {
	return t.options[tagFlatten]
}

// IsOmitEmpty : fill nothing when this field is empty
func (t *Tag) IsOmitEmpty() bool {
	return t.options[tagOmitEmpty]
}

// IsUnique : (sql exclusive)
func (t *Tag) IsUnique() bool {
	return t.options[tagUnique]
}

// IsNullable : (sql exclusive)
func (t *Tag) IsNullable() bool {
	return t.options[tagNullable]
}

// IsUnsigned : (sql exclusive)
func (t *Tag) IsUnsigned() bool {
	return t.options[tagUnsigned]
}

// IsLongText : (sql exclusive)
func (t *Tag) IsLongText() bool {
	return t.options[tagLongText]
}

// newTag :
func newTag(r reflect.StructField) *Tag {
	tagOptions := make([]string, 0)
	tagName := r.Name

	tag := strings.TrimSpace(r.Tag.Get("datastore"))
	paths := strings.Split(tag, ",")
	tagOptions = append(tagOptions, paths[1:]...)
	if strings.TrimSpace(paths[0]) != "" {
		tagName = paths[0]
	}

	tag = strings.TrimSpace(r.Tag.Get("goloquent"))
	paths = strings.Split(tag, ",")
	tagOptions = append(tagOptions, paths[1:]...)
	if strings.TrimSpace(paths[0]) != "" {
		tagName = paths[0]
	}

	options := map[string]bool{
		tagOmitEmpty: false,
		tagNoIndex:   false,
		tagFlatten:   false,
		tagNullable:  false,
		tagUnique:    false, // Extra
		tagUnsigned:  false, // Extra
		tagLongText:  false, // Extra
	}

	// sync tag option
	if len(tagOptions) > 0 {
		for _, name := range tagOptions {
			name = strings.ToLower(name)
			if _, isExist := options[name]; isExist {
				options[name] = true
			}
		}
	}

	t := &Tag{
		Name:         tagName,
		FieldName:    r.Name,
		isPrimaryKey: tagName == "__key__",
		options:      options,
	}

	return t
}
