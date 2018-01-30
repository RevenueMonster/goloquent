package goloquent

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
)

// Connection :
type Connection struct {
	db      string
	adapter Adapter
}

// Close connection
func (c *Connection) Close() {
	fmt.Println("Connection close")
}

// Table :
func (c *Connection) Table(name string) *Table {
	return newTable(name, c)
}

// RunInTransaction :
func (c *Connection) RunInTransaction(callback func(*Connection) error) error {
	return newBuilder(newQuery(newTable("", c))).RunInTransaction(callback)
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
	fmt.Println(connStr)
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

var connPool = map[string]map[string]*Connection{}
var dbTypes = map[string]func(string) (*Connection, error){
	dbDataStore: newDatastore,
	dbMySQL:     newMySQL,
}

// Open : open connection to database
func Open(db string, connString string) (*Connection, error) {
	db = strings.TrimSpace(strings.ToLower(db))
	dbFunc, isValid := dbTypes[db]
	if !isValid {
		panic(ErrUnsupportDatabase)
	}
	pool := make(map[string]*Connection, 0)
	if p, isExist := connPool[db]; isExist {
		pool = p
	}
	conn, err := dbFunc(connString)
	if err != nil {
		return nil, err
	}
	pool[connString] = conn
	connPool[db] = pool
	return conn, nil
}
