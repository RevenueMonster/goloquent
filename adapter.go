package goloquent

import "cloud.google.com/go/datastore"

// Adapter class
type Adapter interface {
	newPrimaryKey(string, *datastore.Key) *datastore.Key
	Find(*Query, *datastore.Key, interface{}) error         // Single record
	First(*Query, interface{}) error                        // Single record
	Get(*Query, interface{}) error                          // Multiple record
	Paginate(*Query, *Pagination, interface{}) error        // Multiple record
	Count(*Query) (int, error)                              // Aggregation
	Create(*Query, interface{}, *datastore.Key) error       // Single record
	Update(*Query, interface{}) error                       // Single record
	Delete(*Query, *datastore.Key) error                    // Single record
	SoftDelete(*Query, *datastore.Key) error                // Single record
	RunInTransaction(*Table, func(*Connection) error) error // Transaction
	CreateMulti(*Query, interface{}, interface{}) error     // Multiple record
	UpsertMulti(*Query, interface{}, interface{}) error     // Multiple record
	UpdateMulti(*Query, interface{}) error                  // Multiple record
	// FindMulti(*Query, []*datastore.Key, interface{}) error // Multiple record
	// DeleteMulti([]*datastore.Key) error // Multiple record
}
