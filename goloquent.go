package goloquent

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"

	"cloud.google.com/go/datastore"
)

// RawQuery :
type RawQuery struct {
	Name string
}

// Connection :
type Connection struct {
	db      string
	adapter Adapter
}

// SetDebug :
func SetDebug(debug bool) {
	isDebug = debug
}

// Close connection
func (c *Connection) Close() error {
	return c.adapter.Close()
}

// Table :
func (c *Connection) Table(name string) *Table {
	return newTable(name, c)
}

// Migrate :
func (c *Connection) Migrate(src interface{}) error {
	adapter, isOk := c.adapter.(*SQLAdapter)
	if !isOk {
		return errors.New("goloquent: not compactible")
	}
	return adapter.Migrate(newQuery(newTable("", c)), src)
}

// Find :
func (c *Connection) Find(key *datastore.Key, src interface{}) error {
	return newBuilder(newQuery(newTable("", c))).Find(key, src)
}

// First :
func (c *Connection) First(src interface{}) error {
	return newBuilder(newQuery(newTable("", c))).First(src)
}

// Get :
func (c *Connection) Get(src interface{}) error {
	return newBuilder(newQuery(newTable("", c))).Get(src)
}

// Paginate :
func (c *Connection) Paginate(p *Pagination, src interface{}) error {
	return newBuilder(newQuery(newTable("", c))).Paginate(p, src)
}

// Create :
func (c *Connection) Create(src interface{}, key interface{}) error {
	return newBuilder(newQuery(newTable("", c))).Create(src, key)
}

// Ancestor :
func (c *Connection) Ancestor(ancestorKey *datastore.Key) *Query {
	return newQuery(newTable("", c)).Ancestor(ancestorKey)
}

// Where :
func (c *Connection) Where(field string, operator string, value interface{}) *Query {
	return newQuery(newTable("", c)).Where(field, operator, value)
}

// Upsert :
func (c *Connection) Upsert(src interface{}, key interface{}) error {
	return newBuilder(newQuery(newTable("", c))).Upsert(src, key)
}

// Delete :
func (c *Connection) Delete(key *datastore.Key) error {
	return newBuilder(newQuery(newTable("", c))).Delete(key)
}

// RunInTransaction :
func (c *Connection) RunInTransaction(callback func(*Connection) error) error {
	return newBuilder(newQuery(newTable("", c))).RunInTransaction(callback)
}

// Statement :
func (c *Connection) Statement(query string, args ...interface{}) ([]map[string][]byte, error) {
	adapter, isOK := c.adapter.(*SQLAdapter)
	if !isOK {
		panic(errors.New("goloquent: unsupported feature"))
	}
	return adapter.ExecQuery(query, args...)
}

func newDatastore(connStr string) (*Connection, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &Connection{
		db: dbDataStore,
		adapter: &DataStoreAdapter{
			client:  client,
			context: ctx,
		},
	}, nil
}

func newMySQL(connStr string) (*Connection, error) {
	client, err := sql.Open(dbMySQL, connStr)
	if err != nil {
		panic(err)
	}
	if e := client.Ping(); e != nil {
		return nil, e
	}
	paths := strings.Split(connStr, "/")
	return &Connection{
		db: dbMySQL,
		adapter: &SQLAdapter{
			mode:   modeNormal,
			client: client,
			dbName: paths[len(paths)-1],
		},
	}, nil
}

var connPool sync.Map
var dbTypes = map[string]func(string) (*Connection, error){
	dbDataStore: newDatastore,
	dbMySQL:     newMySQL,
}

// Raw : raw query
func Raw(field string) *RawQuery {
	return &RawQuery{Name: field}
}

// Open : open connection to database
func Open(db string, connString string) (*Connection, error) {
	db = strings.TrimSpace(strings.ToLower(db))
	dbFunc, isValid := dbTypes[db]
	if !isValid {
		panic(ErrUnsupportDatabase)
	}
	pool := make(map[string]*Connection, 0)
	if p, isExist := connPool.Load(db); isExist {
		pool = p.(map[string]*Connection)
	}
	conn, err := dbFunc(connString)
	if err != nil {
		return nil, err
	}
	pool[connString] = conn
	connPool.Store(db, pool)
	return conn, nil
}
