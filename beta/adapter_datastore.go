package goloquent

import (
	"context"

	"cloud.google.com/go/datastore"
)

// DataStoreAdapter :
type DataStoreAdapter struct {
	client  *datastore.Client
	context context.Context
}

var _ Adapter = &DataStoreAdapter{}

func (ds *DataStoreAdapter) newPrimaryKey(table string, parentKey *datastore.Key) *datastore.Key {
	if parentKey != nil && ((parentKey.Kind == table && parentKey.Name != "") ||
		(parentKey.Kind == table && parentKey.ID > 0)) {
		return parentKey
	}

	return datastore.IncompleteKey(table, parentKey)
}

// CompileQuery :
func (ds *DataStoreAdapter) CompileQuery(query *Query) (*Statement, error) {
	return nil, nil
}

// Close :
func (ds *DataStoreAdapter) Close() error {
	return ds.client.Close()
}
