package v1

// QueryAction :
type QueryAction interface {
	Create(interface{}, interface{}) error
	Upsert(interface{}, interface{}, []string) error
}

// DBQuery :
type DBQuery interface {
	QueryAction
	// Drop() error
	Migrate(...interface{}) error
	RunInTransaction() error
	Statement(string) error
}

// Adapter :
type Adapter interface {
}

// DB :
type DB struct {
	adapter Adapter
}

var _ DBQuery = &DB{}

// Table :
func (db *DB) Table(name string) {

}

// Migrate :
func (db *DB) Migrate(models ...interface{}) error {
	return nil
}

// Create :
func (db *DB) Create(models interface{}, keys interface{}) error {
	return nil
}

// Upsert :
func (db *DB) Upsert(models interface{}, keys interface{}, excludedFields []string) error {
	return nil
}

// RunInTransaction :
func (db *DB) RunInTransaction() error {
	return nil
}

// Statement :
func (db *DB) Statement(sql string) error {
	return nil
}
