package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

// Field :
type Field struct {
	Name         string
	FieldName    string
	IsPrimaryKey bool
	IsOmitEmpty  bool
	Type         reflect.Type
	Index        []int
	Schema       *FieldSchema
	mapper       *Mapper
	// StructField     reflect.StructField
	// Tag             *Tag
}

// String : convert value to string
func (f *Field) String(i interface{}) (*string, error) {
	return f.mapper.StringFunc(i)
}

// Parse : parse string value and convert to relative datatype
func (f *Field) Parse(i string) (interface{}, error) {
	return f.mapper.ParseFunc(f, i)
}

// newField :
func newField(s reflect.StructField, tag *Tag, name []string, index []int, schema *FieldSchema) *Field {
	if schema == nil {
		schema = &FieldSchema{"text", nil, true, false, false, true, utf8CharSet}
	}

	t := s.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	mapper, isValid := mappingList[t.Kind()]
	if !isValid {
		panic(ErrUnsupportDataType)
	}

	// schema.IsNullable = tag.IsNullable()

	f := &Field{
		Name:         strings.Join(name, "."),
		FieldName:    tag.FieldName,
		IsPrimaryKey: tag.IsPrimaryKey(),
		IsOmitEmpty:  tag.IsOmitEmpty(),
		Type:         s.Type,
		Index:        index,
		Schema:       schema,
		mapper:       mapper,
		// Tag:          tag,
		// StructField:  s,
	}

	return f
}

// StructScan :
type StructScan struct {
	Type   reflect.Type
	Column []string
	Index  []int
}

func getSchema(tag *Tag, t reflect.Type) (*FieldSchema, bool) {
	var schema *FieldSchema

	var isPtr = false
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		isPtr = true
	}

	switch t {
	case typeOfString:
		schema = &FieldSchema{fmt.Sprintf("varchar(%d)", TextLength), "", true, tag.IsUnique(), false, tag.IsNullable(), utf8CharSet}
		if tag.IsLongText() {
			schema = &FieldSchema{"text", nil, true, tag.IsUnique(), false, tag.IsNullable(), utf8CharSet}
		}

	case typeOfBool:
		schema = &FieldSchema{"boolean", false, false, false, false, false, nil}

	case typeOfInt, typeOfInt8, typeOfInt16, typeOfInt32:
		schema = &FieldSchema{"int", int(0), false, tag.IsUnique(), tag.IsUnsigned(), false, nil}

	case typeOfInt64:
		schema = &FieldSchema{"bigint", int64(0), false, tag.IsUnique(), tag.IsUnsigned(), false, nil}

	case typeOfFloat32:
		schema = &FieldSchema{"double(8,6)", float32(0), false, tag.IsUnique(), tag.IsUnsigned(), false, nil}

	case typeOfFloat64:
		schema = &FieldSchema{"decimal(10,6)", float64(0), false, tag.IsUnique(), tag.IsUnsigned(), false, nil}

	case typeOfByte:
		schema = &FieldSchema{"mediumblob", nil, true, tag.IsUnique(), false, false, nil}

	case typeOfSoftDelete:
		schema = &FieldSchema{"datetime", nil, true, false, false, true, nil}

	case typeOfTime:
		schema = &FieldSchema{"datetime", time.Time{}, true, false, false, false, nil}

	case typeOfDataStoreKey:
		schema = &FieldSchema{fmt.Sprintf("varchar(%d)", KeyLength), nil, true, tag.IsUnique(), false, true, latin2CharSet}

	case typeOfGeopoint:
		schema = &FieldSchema{"varchar(50)", datastore.GeoPoint{}, true, false, false, isPtr, latin2CharSet}

	default:
		// slice, array or struct will be text
		return &FieldSchema{"text", nil, true, tag.IsUnique(), false, false, utf8CharSet}, false

	}

	return schema, true
}

// ListFields :
func ListFields(t reflect.Type) (*Field, *Field, []*Field, error) {

	var pk, soft *Field
	fields := make([]*Field, 0)
	scanStructs := make([]*StructScan, 0)
	scanStructs = append(scanStructs, &StructScan{
		Column: make([]string, 0),
		Type:   t,
		Index:  make([]int, 0),
	})

	for len(scanStructs) > 0 {
		// begin with the first struct
		r := scanStructs[0]
		for i := 0; i < r.Type.NumField(); i++ {
			f := r.Type.Field(i)
			t := f.Type

			isExported := (f.PkgPath == "")
			// Skip if not anonymous private property
			if !isExported && !f.Anonymous {
				continue
			}

			tag := newTag(f)
			// Skip if contain skip tag
			if tag.IsSkip() {
				continue
			}

			if isNameReserved(tag.Name) {
				return nil, nil, nil, fmt.Errorf("goloquent: name `%s` is reserved", tag.Name)
			}

			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}

			col := make([]string, 0)
			col = append(col, r.Column...)
			col = append(col, tag.Name)

			index := make([]int, 0)
			index = append(index, r.Index...)
			index = append(index, i)

			if tag.IsPrimaryKey() {
				if t != typeOfDataStoreKey {
					return nil, nil, nil, errors.New("goloquent: invalid datatype of primary key")
				}
				pk = newField(f, tag, col, index, nil)
				continue
			}

			if t == typeOfSoftDelete {
				schema, _ := getSchema(tag, t)
				soft = newField(f, tag, []string{FieldNameSoftDelete}, index, schema)
				continue
			}

			// if it's flatten just continue on next
			if tag.IsFlatten() {
				if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
					t = t.Elem()
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					if _, isLeaf := getSchema(tag, t); isLeaf {
						fields = append(fields, newField(f, tag, col, index, nil))
						continue
					}
				}

				scanStructs = append(scanStructs, &StructScan{
					Column: col,
					Type:   t,
					Index:  index,
				})
				continue
			}

			if schema, isLeaf := getSchema(tag, t); isLeaf {
				fields = append(fields, newField(f, tag, col, index, schema))
				continue
			}

			fs := newField(f, tag, col, index, nil)

			// if it's embedded struct
			if f.Anonymous {
				// Skip if embedded struct is unexported
				if !isExported {
					continue
				}
				scanStructs = append(scanStructs, &StructScan{
					Column: make([]string, 0),
					Type:   t,
					Index:  index,
				})
				continue
			}

			fields = append(fields, fs)
		}

		// unshift scan struct
		scanStructs = scanStructs[1:]
	}

	// Sort the column according to the sequence of model
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Index[0] < fields[j].Index[0]
	})

	return pk, soft, fields, nil
}
