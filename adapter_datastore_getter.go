package goloquent

import (
	"fmt"

	"cloud.google.com/go/datastore"
)

// Find :
func (ds *DataStoreAdapter) Find(q *Query, key *datastore.Key, modelStruct interface{}) error {
	return ds.client.Get(ds.context, key, modelStruct)
}

// First :
func (ds *DataStoreAdapter) First(q *Query, modelStruct interface{}) error {
	fmt.Println("datastore first")
	return nil
}

// Get :
func (ds *DataStoreAdapter) Get(query *Query, modelStruct interface{}) error {
	q, err := ds.CompileQuery(query)
	if err != nil {
		return err
	}

	if _, err := ds.client.GetAll(ds.context, q, modelStruct); err != nil {
		return err
	}

	return nil
}

// Paginate :
func (ds *DataStoreAdapter) Paginate(query *Query, p *Pagination, modelStruct interface{}) error {
	fmt.Println("datastore paginate")
	return nil
}
