package goloquent

import (
	"cloud.google.com/go/datastore"
)

// Table :
type Table struct {
	name       string
	connection *Connection
}

func newTable(name string, c *Connection) *Table {
	return &Table{
		name:       name,
		connection: c,
	}
}

func (t *Table) getAdapter() Adapter {
	return t.connection.adapter
}

func (t *Table) getSQLAdapter() (*SQLAdapter, error) {
	adapter, isOK := t.getAdapter().(*SQLAdapter)
	if !isOK {
		return nil, ErrUnsupportFeature
	}
	return adapter, nil
}

// Find :
func (t *Table) Find(key *datastore.Key, modelStruct interface{}) error {
	return newBuilder(newQuery(t)).Find(key, modelStruct)
}

// First :
func (t *Table) First(modelStruct interface{}) error {
	return newBuilder(newQuery(t)).First(modelStruct)
}

// Get :
func (t *Table) Get(modelStruct interface{}) error {
	return newBuilder(newQuery(t)).Get(modelStruct)
}

// Paginate :
func (t *Table) Paginate(p *Pagination, modelStruct interface{}) error {
	return newBuilder(newQuery(t)).Paginate(p, modelStruct)
}

// Ancestor :
func (t *Table) Ancestor(parentKey *datastore.Key) *Query {
	return newQuery(t).Ancestor(parentKey)
}

// NewQuery :
func (t *Table) NewQuery() *Query {
	return newQuery(t)
}

// Where :
func (t *Table) Where(field string, o string, value interface{}) *Query {
	return newQuery(t).Where(field, o, value)
}

// // WhereNotLike :
// func (t *Table) WhereNotLike(field string, value interface{}) *Query {
// 	return newQuery(t).WhereNotLike(field, value)
// }

// // WhereLike :
// func (t *Table) WhereLike(field string, value interface{}) *Query {
// 	return newQuery(t).WhereLike(field, value)
// }

// WithTrashed :
func (t *Table) WithTrashed() *Query {
	return newQuery(t).WithTrashed()
}

// Order :
func (t *Table) Order(fields string) *Query {
	return newQuery(t).Order(fields)
}

// Limit :
func (t *Table) Limit(i int) *Limit {
	return newQuery(t).Limit(i)
}

// Count :
func (t *Table) Count() (uint, error) {
	return newBuilder(newQuery(t)).Count()
}

// Create :
func (t *Table) Create(modelStruct interface{}, parentKey interface{}) error {
	return newBuilder(newQuery(t)).Create(modelStruct, parentKey)
}

// Upsert :
func (t *Table) Upsert(modelStruct interface{}, parentKey interface{}) error {
	return newBuilder(newQuery(t)).Upsert(modelStruct, parentKey)
}

// Update :
func (t *Table) Update(modelStruct interface{}) error {
	return newBuilder(newQuery(t)).Update(modelStruct)
}

// Delete :
func (t *Table) Delete(key *datastore.Key) error {
	return newBuilder(newQuery(t)).Delete(key)
}

// SoftDelete :
func (t *Table) SoftDelete(key *datastore.Key) error {
	return newBuilder(newQuery(t)).SoftDelete(key)
}

// LockForShared :
func (t *Table) LockForShared() *Getter {
	return newQuery(t).LockForShared()
}

// LockForUpdate :
func (t *Table) LockForUpdate() *Getter {
	return newQuery(t).LockForUpdate()
}

// Migrate : (SQL exclusive actions)
func (t *Table) Migrate(modelStruct interface{}) error {
	adapter, err := t.getSQLAdapter()
	if err != nil {
		return err
	}
	return adapter.Migrate(newQuery(t), modelStruct)
}

// Drop : (SQL exclusive actions)
func (t *Table) Drop() error {
	adapter, err := t.getSQLAdapter()
	if err != nil {
		return err
	}
	return adapter.Drop(newQuery(t))
}

// DropIfExists : (SQL exclusive actions)
func (t *Table) DropIfExists() error {
	adapter, err := t.getSQLAdapter()
	if err != nil {
		return err
	}
	return adapter.DropIfExists(newQuery(t))
}

// UniqueIndex : (SQL exclusive actions)
func (t *Table) UniqueIndex(fields ...string) error {
	adapter, err := t.getSQLAdapter()
	if err != nil {
		return err
	}
	return adapter.UniqueIndex(newQuery(t), fields...)
}

// Sum :
func (t *Table) Sum(field string) (uint, error) {
	adapter, err := t.getSQLAdapter()
	if err != nil {
		return 0, err
	}
	return adapter.Sum(field, newQuery(t))
}
