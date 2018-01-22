package goloquent

import (
	"fmt"

	"cloud.google.com/go/datastore"
)

// Find :
func (ds *DataStoreAdapter) Find(q *Query, key *datastore.Key, modelStruct interface{}) error {
	fmt.Println("datastore find")
	if err := ds.client.Get(ds.context, key, modelStruct); err != nil {
		return err
	}
	return nil
}

// First :
func (ds *DataStoreAdapter) First(q *Query, modelStruct interface{}) error {
	fmt.Println("datastore first")
	return nil
}

// Get :
func (ds *DataStoreAdapter) Get(q *Query, modelStruct interface{}) error {
	fmt.Println("datastore get")
	return nil
}

// Paginate :
func (ds *DataStoreAdapter) Paginate(query *Query, p *Pagination, modelStruct interface{}) error {
	fmt.Println("datastore paginate")
	return nil
}
