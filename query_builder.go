package goloquent

import (
	"errors"
	"fmt"
	"reflect"

	"cloud.google.com/go/datastore"
)

// Builder :
type Builder struct {
	query *Query
}

func newBuilder(q *Query) *Builder {
	return &Builder{
		query: q,
	}
}

func (b *Builder) getAdapter() Adapter {
	return b.query.table.connection.adapter
}

// Find :
func (b *Builder) Find(key *datastore.Key, modelStruct interface{}) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	typeOf := reflect.TypeOf(modelStruct)
	if typeOf.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}
	e := typeOf.Elem()
	if e.Kind() != reflect.Struct {
		return errors.New("goloquent: model must be struct")
	}

	return b.query.table.connection.adapter.Find(b.query, key, modelStruct)
}

// Get :
func (b *Builder) Get(modelStruct interface{}) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	typeOf := reflect.TypeOf(modelStruct)
	if typeOf.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}

	return b.getAdapter().Get(b.query, modelStruct)
}

// First :
func (b *Builder) First(modelStruct interface{}) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	t := reflect.TypeOf(modelStruct)
	if t.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}

	return b.getAdapter().First(b.query, modelStruct)
}

// Paginate :
func (b *Builder) Paginate(p *Pagination, modelStruct interface{}) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	switch val := p.Filter.(type) {
	case nil:
	case []Filter:
		b.query.filters = append(b.query.filters, val...)

	default:
		filters := make([]Filter, 0)
		v := reflect.Indirect(reflect.ValueOf(p.Filter))
		switch v.Kind() {
		case reflect.Struct:
			_, _, fields, err := ListFields(v.Type())
			if err != nil {
				return err
			}

			for _, item := range fields {
				f := v.FieldByIndex(item.Index)
				//  || isZero(f.Interface())
				if !f.IsValid() {
					continue
				}

				filter := newFilter(item.Name, "=", f.Interface())
				filters = append(filters, filter)
			}

		// TODO: support for map filter
		// case reflect.Map:
		// for _, key := range v.MapKeys() {
		// 	f := v.MapIndex(key)
		// 	t := reflect.TypeOf(f.Interface())
		// 	if t.Kind() == reflect.Ptr {
		// 		t = t.Elem()
		// 	}
		// 	filters = append(filters, newFilter(t.Name(), "=", f.Interface(), operators["="]))
		// }

		default:
			return errors.New("goloquent: invalid paginate filter datatype")
		}

		b.query.filters = append(b.query.filters, filters...)
	}

	if len(p.OrderBy) > 0 {
		b.query.orders = append(b.query.orders, p.OrderBy...)
	}

	return b.getAdapter().Paginate(b.query, p, modelStruct)
}

// Count :
func (b *Builder) Count() (uint, error) {
	if len(b.query.errs) > 0 {
		return 0, b.query.errs[0]
	}

	return b.getAdapter().Count(b.query)
}

// Create :
func (b *Builder) Create(modelStruct interface{}, parentKey interface{}) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	t := reflect.TypeOf(modelStruct)
	if t.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}
	t = t.Elem()
	if t.Kind() == reflect.Struct {
		if parentKey == nil {
			return b.getAdapter().Create(b.query, modelStruct, nil)
		}
		key, isValid := parentKey.(*datastore.Key)
		if !isValid {
			return errors.New("goloquent: invalid key datatype")
		}
		return b.getAdapter().Create(b.query, modelStruct, key)
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	if v.Len() <= 0 {
		return errors.New("goloquent: no valid record to insert")
	}
	if v.Len() > int(MaxRecordInsert) {
		return fmt.Errorf("goloquent: maximum insert records, %d", MaxRecordInsert)
	}

	return b.getAdapter().CreateMulti(b.query, modelStruct, parentKey)
}

// Upsert :
func (b *Builder) Upsert(modelStruct interface{}, parentKey interface{}, excluded ...string) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	t := reflect.TypeOf(modelStruct)
	if t.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}

	t = t.Elem()
	if t.Kind() == reflect.Struct {
		if parentKey == nil {
			return b.getAdapter().Upsert(b.query, modelStruct, nil, excluded...)
		}
		key, isValid := parentKey.(*datastore.Key)
		if !isValid {
			return errors.New("goloquent: invalid key datatype")
		}
		return b.getAdapter().Upsert(b.query, modelStruct, key, excluded...)
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	if v.Len() <= 0 {
		return errors.New("goloquent: no valid record to insert")
	}
	if v.Len() > int(MaxRecordInsert) {
		return fmt.Errorf("goloquent: maximum insert records, %d", MaxRecordInsert)
	}

	return b.getAdapter().UpsertMulti(b.query, modelStruct, parentKey, excluded...)
}

// Update :
func (b *Builder) Update(modelStruct interface{}) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	t := reflect.TypeOf(modelStruct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return errors.New("goloquent: invalid model")
	}

	return b.getAdapter().Update(b.query, modelStruct)
}

// UpdateMulti :
func (b *Builder) UpdateMulti(modelStruct interface{}) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	if v.Kind() != reflect.Map && v.Kind() != reflect.Struct {
		return fmt.Errorf(
			"goloquent: invalid data type %T on update multiple records", modelStruct)
	}

	return b.getAdapter().UpdateMulti(b.query, modelStruct)
}

// Delete :
func (b *Builder) Delete(key *datastore.Key) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	if key == nil {
		return errors.New("goloquent: datastore key cannot be nil")
	}
	if key.Incomplete() {
		return errors.New("goloquent: datastore key is incomplete")
	}
	return b.getAdapter().Delete(b.query, key)
}

// SoftDelete :
func (b *Builder) SoftDelete(key *datastore.Key) error {
	if len(b.query.errs) > 0 {
		return b.query.errs[0]
	}

	if key == nil {
		return errors.New("goloquent: datastore key cannot be nil")
	}
	if key.Incomplete() {
		return errors.New("goloquent: datastore key is incomplete")
	}
	return b.getAdapter().SoftDelete(b.query, key)
}

// RunInTransaction :
func (b *Builder) RunInTransaction(cb func(*Connection) error) error {
	return b.getAdapter().RunInTransaction(b.query.table, cb)
}
