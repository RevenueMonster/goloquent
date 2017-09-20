package db

import (
	"context"
	"errors"
	"reflect"

	"google.golang.org/api/iterator"

	"fmt"

	"cloud.google.com/go/datastore"
)

// DataStoreClient is the global google cloud data client
var (
	DataStoreClient *datastore.Client
	Ctx             = context.Background()
)

// Config :
func Config(projectID string) error {
	var err error
	DataStoreClient, err = datastore.NewClient(Ctx, projectID)
	if err != nil {
		return err
	}
	return nil
}

// Pagination :
type Pagination struct {
	Cursor  string      `json:"cursor" validate:"omitempty"`
	Filter  interface{} `json:"filter" validate:"omitempty"`
	OrderBy string      `json:"order" validate:"omitempty"`
	Count   int         `json:"-"`
	Limit   int         `json:"-"`
}

// Eloquent :
type Eloquent struct {
	trx        *datastore.Transaction
	query      *datastore.Query
	result     interface{}
	collection string
	cursor     string
	err        []error
	page       *Pagination
}

// Kind : select collection
func Kind(name string) *Eloquent {
	var errs []error
	return &Eloquent{
		trx:        nil,
		query:      datastore.NewQuery(name),
		result:     nil,
		collection: name,
		err:        errs,
		page:       nil,
	}
}

// BeginTransaction :
func BeginTransaction() *Eloquent {
	trx, err := DataStoreClient.NewTransaction(Ctx)
	if err != nil {
		panic("Transaction cannot begin!!!")
	}
	var errs []error
	return &Eloquent{
		trx: trx,
		err: errs,
	}
}

// Kind : select collection
func (e *Eloquent) Kind(name string) *Eloquent {
	e.collection = name
	e.query = datastore.NewQuery(name)
	return e
}

// Commit : Commit the transaction
func (e *Eloquent) Commit() error {
	_, err := e.trx.Commit()
	return err
}

// Rollback : Rollback the transaction
func (e *Eloquent) Rollback() error {
	err := e.trx.Rollback()
	return err
}

// Create : Create a single record
func (e *Eloquent) Create(modelStruct interface{}, parentKey *datastore.Key) error {
	var key *datastore.Key
	if parentKey != nil && ((parentKey.Kind == e.collection && parentKey.Name != "") ||
		(parentKey.Kind == e.collection && parentKey.ID > 0)) {
		key = parentKey
	} else {
		key = datastore.IncompleteKey(e.collection, parentKey)
	}

	if e.trx != nil {
		_, err := e.trx.Put(key, modelStruct)
		if err != nil {
			return err
		}
	} else {
		k, err := DataStoreClient.Put(Ctx, key, modelStruct)
		if err != nil {
			return err
		}
		v := reflect.ValueOf(modelStruct)
		f := v.Elem().FieldByName("Key")
		if f.Kind() == reflect.Ptr && f.CanSet() == true {
			f.Set(reflect.ValueOf(k))
		}
	}
	return nil
}

// CreateMulti : Create a multiple record
func (e *Eloquent) CreateMulti(modelStruct interface{}, parentKey *datastore.Key) error {
	v := reflect.ValueOf(modelStruct)
	var pKeys []*datastore.Key
	m := make([]interface{}, v.Len())

	for i := 0; i < v.Len(); i++ {
		m[i] = v.Index(i).Interface()
		k := new(datastore.Key)
		k = datastore.IncompleteKey(e.collection, parentKey)

		pKeys = append(pKeys, k)
	}

	if e.trx != nil {
		_, err := e.trx.PutMulti(pKeys, modelStruct)
		if err != nil {
			return err
		}
	} else {
		keys, err := DataStoreClient.PutMulti(Ctx, pKeys, m)
		if err != nil {
			return err
		}

		for index := range keys {
			v := reflect.ValueOf(m[index])
			f := v.Elem().FieldByName("Key")
			if f.Kind() == reflect.Ptr && f.CanSet() == true {
				f.Set(reflect.ValueOf(keys[index]))
			}
		}
	}
	return nil
}

// Where : Filter by field
func (e *Eloquent) Where(field string, value interface{}) *Eloquent {
	e.query = e.query.Filter(field, value)
	return e
}

// Find : Find the row using key
func (e *Eloquent) Find(key *datastore.Key, modelStruct interface{}) error {
	if e.trx != nil {
		return e.trx.Get(key, modelStruct)
	}
	return DataStoreClient.Get(Ctx, key, modelStruct)
}

// FindMulti : Find multiple row by using multiple key
func (e *Eloquent) FindMulti(keys []*datastore.Key, modelStruct interface{}) error {
	if e.trx != nil {
		return e.trx.GetMulti(keys, modelStruct)
	}
	return DataStoreClient.GetMulti(Ctx, keys, modelStruct)
}

// First : Get the first record of the query
func (e *Eloquent) First(modelStruct interface{}) error {
	t := reflect.TypeOf(modelStruct)
	s := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	m := reflect.New(s.Type())
	m.Elem().Set(s)

	key, err := DataStoreClient.GetAll(Ctx, e.query.Limit(1), m.Interface())
	if err != nil {
		return err
	}

	if len(key) == 1 {
		j := m.Elem().Index(0).Interface()
		r := reflect.ValueOf(j).Elem()
		g := reflect.ValueOf(modelStruct).Elem()
		for i := 0; i < g.NumField(); i++ {
			g.Field(i).Set(r.Field(i))
		}
		if e.err != nil {
			return e.err[0]
		}
	} else if len(key) > 1 {
		return errors.New("More than one unique")
	}
	return nil
}

// Get : Get multiple record of the query
func (e *Eloquent) Get(modelStruct interface{}) error {
	m := reflect.ValueOf(modelStruct).Elem()

	rm := reflect.TypeOf(modelStruct).Elem().Elem()
	i := reflect.New(rm)
	it := DataStoreClient.Run(Ctx, e.query)
	_, err := it.Next(i.Interface())
	for err == nil {
		m = reflect.Append(m, i.Elem())
		_, err = it.Next(i.Interface())
	}

	if err != iterator.Done {
		return err
	}

	if e.page != nil {
		k, _ := it.Cursor()
		e.page.Cursor = k.String()
	}

	g := reflect.ValueOf(modelStruct).Elem()
	g.Set(m)

	return nil
}

// Count : Count the number of data
func (e *Eloquent) Count() (int, error) {
	n, err := DataStoreClient.Count(Ctx, e.query)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Ancestor :
func (e *Eloquent) Ancestor(key *datastore.Key) *Eloquent {
	// e.ancestors = append(e.ancestors, key)
	e.query = e.query.Ancestor(key)
	return e
}

// Limit :
func (e *Eloquent) Limit(n int) *Eloquent {
	e.query = e.query.Limit(n)
	return e
}

// KeysOnly :
func (e *Eloquent) KeysOnly() *Eloquent {
	e.query = e.query.KeysOnly()
	return e
}

// Paginate :
func (e *Eloquent) Paginate(p *Pagination) *Eloquent {
	e.page = p
	limit := p.Limit

	c, err := DataStoreClient.Count(Ctx, datastore.NewQuery(e.collection))
	if err != nil {
		e.err = append(e.err, err)
	}
	// order := "__key__"
	// if p.OrderBy != "" {
	//  order = p.OrderBy
	// }

	q := e.query
	if reflect.TypeOf(p.Filter) != nil {
		v := reflect.ValueOf(p.Filter)
		t := v.Elem()
		for i := 0; i < t.NumField(); i++ {
			name := reflect.Indirect(v).Type().Field(i).Name
			val := t.Field(i).Interface()
			if val != "" && val != nil && val != 0 {
				q = q.Filter(fmt.Sprintf("%s =", name), t.Field(i).Interface())
			}
		}
	}
	p.Count = c

	if len(p.Cursor) > 0 {
		cursor, err := datastore.DecodeCursor(p.Cursor)
		if err != nil {
			p.Cursor = ""
			e.query = q.Limit(limit)
		} else {
			e.query = q.Limit(limit).Start(cursor)
		}
	} else {
		e.query = q.Limit(limit)
	}

	return e
}

// Update : update a record in db
func (e *Eloquent) Update(modelStruct interface{}) error {
	v := reflect.ValueOf(modelStruct)
	Key := v.Elem().FieldByName("Key").Interface().(*datastore.Key)
	var err error
	if e.trx != nil {
		_, err = e.trx.Put(Key, modelStruct)
	} else {
		_, err = DataStoreClient.Put(Ctx, Key, modelStruct)
	}
	if err != nil {
		return err
	}
	return nil
}

// UpdateMulti : update a record in db
func (e *Eloquent) UpdateMulti(modelStruct interface{}) error {
	v := reflect.ValueOf(modelStruct)
	var pKeys []*datastore.Key
	for i := 0; i < v.Len(); i++ {
		k := v.Index(i).FieldByName("Key").Interface().(*datastore.Key)
		pKeys = append(pKeys, k)
	}

	var err error
	if e.trx != nil {
		_, err = e.trx.PutMulti(pKeys, modelStruct)
	} else {
		_, err = DataStoreClient.PutMulti(Ctx, pKeys, modelStruct)
	}
	if err != nil {
		return err
	}
	return nil
}

// Delete : delete a record in db
func (e *Eloquent) Delete(key *datastore.Key) error {
	if e.trx != nil {
		return e.trx.Delete(key)
	}
	return DataStoreClient.Delete(Ctx, key)
}

// DeleteMulti : delete multiple record in db
func (e *Eloquent) DeleteMulti(key []*datastore.Key) error {
	if e.trx != nil {
		return e.trx.DeleteMulti(key)
	}
	return DataStoreClient.DeleteMulti(Ctx, key)
}
