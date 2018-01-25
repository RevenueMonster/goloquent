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
	typeOf := reflect.TypeOf(modelStruct)
	if typeOf.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}
	e := typeOf.Elem()
	if e.Kind() != reflect.Struct {
		return errors.New("model must be struct")
	}

	return b.query.table.connection.adapter.Find(b.query, key, modelStruct)
}

// Get :
func (b *Builder) Get(modelStruct interface{}) error {
	typeOf := reflect.TypeOf(modelStruct)
	if typeOf.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}
	return b.getAdapter().Get(b.query, modelStruct)
}

// First :
func (b *Builder) First(modelStruct interface{}) error {
	t := reflect.TypeOf(modelStruct)
	if t.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}

	return b.getAdapter().First(b.query, modelStruct)
}

// Paginate :
func (b *Builder) Paginate(p *Pagination, modelStruct interface{}) error {
	return b.getAdapter().Paginate(b.query, p, modelStruct)
}

// Count :
func (b *Builder) Count() (int, error) {
	return b.getAdapter().Count(b.query)
}

// Create :
func (b *Builder) Create(modelStruct interface{}, parentKey interface{}) error {
	t := reflect.TypeOf(modelStruct)
	if t.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}
	t = t.Elem()
	if t.Kind() == reflect.Struct {
		if reflect.TypeOf(parentKey) != typeOfPtrDataStoreKey {
			return errors.New("goloquent: invalid key data type")
		}
		key := parentKey.(*datastore.Key)
		return b.getAdapter().Create(b.query, modelStruct, key)
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	if v.Len() <= 0 {
		return errors.New("goloquent: no valid record to insert")
	}
	if v.Len() > int(MaxRecord) {
		return fmt.Errorf("goloquent: maximum insert records, %d", MaxRecord)
	}

	return b.getAdapter().CreateMulti(b.query, modelStruct, parentKey)
}

// Upsert :
func (b *Builder) Upsert(modelStruct interface{}, parentKey interface{}) error {
	t := reflect.TypeOf(modelStruct)
	if t.Kind() != reflect.Ptr {
		return ErrInvalidDataTypeModel
	}
	t = t.Elem()
	if t.Kind() == reflect.Struct {
		return b.getAdapter().Upsert(b.query, modelStruct, parentKey.(*datastore.Key))
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	if v.Len() <= 0 {
		return errors.New("goloquent: no valid record to insert")
	}
	if v.Len() > int(MaxRecord) {
		return fmt.Errorf("goloquent: maximum insert records, %d", MaxRecord)
	}

	return b.getAdapter().UpsertMulti(b.query, modelStruct, parentKey)
}

// Update :
func (b *Builder) Update(modelStruct interface{}) error {
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
	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	if v.Len() > int(MaxRecord) {
		return fmt.Errorf("goloquent: maximum update records, %d", MaxRecord)
	}

	return b.getAdapter().UpdateMulti(b.query, modelStruct)
}

// Delete :
func (b *Builder) Delete(key *datastore.Key) error {
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
