package goloquent

import (
	"reflect"
	"strings"
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
	tagName := r.Name
	tag := strings.TrimSpace(r.Tag.Get(optionTagDatastore))
	pkgTag := strings.TrimSpace(r.Tag.Get(optionTagGoloquent))

	// Indentify primary key with __key__
	isMatch := IsPrimaryKey(tag)
	paths := strings.Split(tag, ",")
	if strings.TrimSpace(paths[0]) != "" {
		tagName = paths[0]
	}

	t := &Tag{
		Name:         tagName,
		FieldName:    r.Name,
		isPrimaryKey: isMatch,
		options: map[string]bool{
			tagOmitEmpty: false,
			tagNoIndex:   false,
			tagFlatten:   false,
			tagNullable:  false,
			tagUnique:    false, // Extra
			tagUnsigned:  false, // Extra
			tagLongText:  false, // Extra
		},
	}

	// sync tag option
	optionPaths := paths[1:]
	optionPaths = append(optionPaths, strings.Split(pkgTag, ",")...)
	if len(optionPaths) > 0 {
		for _, name := range optionPaths {
			name = strings.ToLower(name)
			if _, isExist := t.options[name]; isExist {
				t.options[name] = true
			}
		}
	}

	return t
}
