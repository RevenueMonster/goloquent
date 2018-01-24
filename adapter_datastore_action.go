package goloquent

import (
	"fmt"

	"cloud.google.com/go/datastore"
)

// Create :
func (ds *DataStoreAdapter) Create(query *Query, modelStruct interface{}, parentKey *datastore.Key) error {
	key := ds.newPrimaryKey(query.table.name, parentKey)
	k, err := ds.client.Put(ds.context, key, modelStruct)
	if err != nil {
		return err
	}
	fmt.Println(k)
	return nil
}

// CreateMulti :
func (ds *DataStoreAdapter) CreateMulti(query *Query, modelStruct interface{}, parentKey interface{}) error {
	return nil
}

// Upsert :
func (ds *DataStoreAdapter) Upsert(query *Query, modelStruct interface{}, parentKey *datastore.Key) error {
	return nil
}

// UpsertMulti :
func (ds *DataStoreAdapter) UpsertMulti(query *Query, modelStruct interface{}, parentKey interface{}) error {
	return nil
}

// Update :
func (ds *DataStoreAdapter) Update(query *Query, modelStruct interface{}) error {
	fmt.Println("datastore update")
	return nil
}

// UpdateMulti :
func (ds *DataStoreAdapter) UpdateMulti(query *Query, modelStruct interface{}) error {
	return nil
}

// Delete :
func (ds *DataStoreAdapter) Delete(query *Query, key *datastore.Key) error {
	fmt.Println("datastore delete")
	return nil
}

// SoftDelete :
func (ds *DataStoreAdapter) SoftDelete(query *Query, key *datastore.Key) error {
	fmt.Println("datastore soft delete")
	return nil
}

// RunInTransaction :
func (ds *DataStoreAdapter) RunInTransaction(table *Table, callback func(*Connection) error) error {
	fmt.Println("run in transaction")
	return nil
}
