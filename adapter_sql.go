package goloquent

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/fatih/color"
)

// SQLAdapter :
type SQLAdapter struct {
	table  string
	mode   string
	dbName string
	client *sql.DB
	txn    *sql.Tx
}

// Statement :
type Statement struct {
	Where  []string
	Order  []string
	Limit  uint
	Locked string
}

var _ Adapter = &SQLAdapter{}

func (x *SQLAdapter) newPrimaryKey(table string, parentKey *datastore.Key) *datastore.Key {
	if parentKey != nil && ((parentKey.Kind == table && parentKey.Name != "") ||
		(parentKey.Kind == table && parentKey.ID > 0)) {
		return parentKey
	}

	key := new(datastore.Key)
	rand.Seed(time.Now().UnixNano())
	id := rand.Int63n(MaxSeed-MinSeed) + MinSeed
	key.Kind = table
	key.ID = id
	if parentKey != nil {
		key.Parent = parentKey
	}
	return key
}

func (x *SQLAdapter) sqlDebug(sql string) {
	if isDebug {
		fmt.Println("************* START QUERY STATEMENT ************")
		fmt.Println(color.GreenString(sql))
		fmt.Println("************* ENDED QUERY STATEMENT ************")
	}
}

func (x *SQLAdapter) mapResults(query *Query, e *Entity, t reflect.Type, results []map[string][]byte) (reflect.Value, error) {
	slice := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	isPtr := (t.Kind() == reflect.Ptr)
	if isPtr {
		t = t.Elem()
	}

	table := query.table.name
	fields := e.GetFields()
	// Run through every record
	for _, rec := range results {
		i := reflect.New(t)
		for _, fs := range fields {
			f := i.Elem().FieldByIndex(fs.Index)
			if !f.IsValid() {
				continue
			}

			b, isExist := rec[fs.Name]
			if !isExist {
				continue
			}

			it, err := fs.Parse(string(b))
			if err != nil {
				return slice, err
			}

			iv := reflect.ValueOf(it)
			if iv.IsValid() {
				f.Set(iv)
			}
		}

		if err := e.LoadFunc(i.Interface()); err != nil {
			return slice, err
		}

		// Load PrimaryKey
		pk := rec[FieldNameKey]
		parent := rec[FieldNameParent]
		primaryKey, err := parsePrimaryKey(table, string(pk), string(parent))
		if err != nil {
			return slice, err
		}

		if err := e.LoadKey(i.Interface(), primaryKey); err != nil {
			return slice, err
		}

		if !isPtr {
			i = i.Elem()
		}

		slice = reflect.Append(slice, i)

	}

	return slice, nil
}

// Exec :
func (x *SQLAdapter) Exec(q string) (sql.Result, error) {
	go x.sqlDebug(q)

	if x.mode == modeNormal {
		return x.client.Exec(q)
	}

	return x.txn.Exec(q)
}

// ExecQuery :
func (x *SQLAdapter) ExecQuery(q string, args ...interface{}) ([]map[string][]byte, error) {
	r := make([]map[string][]byte, 0)

	go func(query string) {
		sql := query
		for i := range args {
			s, isOK := args[i].(string)
			if isOK {
				sql = strings.Replace(sql, "?", fmt.Sprintf("%q", s), 1)
			}
		}
		x.sqlDebug(sql)
	}(q)

	var (
		rows *sql.Rows
		err  error
	)

	if x.mode == modeNormal {
		rows, err = x.client.Query(q, args...)
	} else {
		rows, err = x.txn.Query(q, args...)
	}

	if err != nil {
		return r, err
	}

	defer rows.Close()

	c, _ := rows.Columns()
	intCols := len(c)

	for rows.Next() {
		m := make([]interface{}, intCols)

		for i := range c {
			m[i] = &m[i]
		}

		if err := rows.Scan(m...); err != nil {
			return r, err
		}

		rows := make(map[string][]byte)
		for i, key := range c {
			var (
				b    []byte
				isOK bool
			)
			if m[i] == nil {
				b = []byte("")
			} else {
				b, isOK = m[i].([]byte)
				if !isOK {
					b = []byte(m[i].(string))
				}
			}
			rows[key] = b
		}

		r = append(r, rows)
	}

	return r, nil
}

// CompileStatement :
func (x *SQLAdapter) CompileStatement(query *Query) (*Statement, error) {
	where := make([]string, 0)
	if len(query.filters) > 0 {
		for _, f := range query.filters {
			if f.Field == tagKey {
				v := reflect.ValueOf(f.Value)
				q := ""
				switch v.Kind() {
				case reflect.Slice, reflect.Array:
					strKeys := make([]string, 0)
					for i := 0; i < v.Len(); i++ {
						k := v.Index(i).Interface().(*datastore.Key)
						strKeys = append(strKeys,
							fmt.Sprintf("%q", k.Parent.String()+"/"+stringPrimaryKey(k)))
					}
					q = fmt.Sprintf(
						"CONCAT(`%s`,%q,`%s`) IN (%s)",
						FieldNameParent, "/", FieldNameKey,
						strings.Join(strKeys, ","))

				default:
					k := f.Value.(*datastore.Key)
					strKey := stringPrimaryKey(k)
					q = fmt.Sprintf("(`%s` = %q AND `%s` = %q)",
						FieldNameKey, strKey, FieldNameParent, k.Parent.String())
				}
				where = append(where, q)
				continue
			}

			s, err := f.String()
			if err != nil {
				return nil, err
			}

			strVal := ""
			if s == nil {
				if f.Operator == "=" {
					strVal = "IS NULL"
				} else if f.Operator == "!=" {
					strVal = "IS NOT NULL"
				}
			} else {
				strVal = fmt.Sprintf("%s %s", f.Operator, *s)
			}

			if strVal == "" {
				return nil, errors.New("goloquent: invalid datatype")
			}

			where = append(where, fmt.Sprintf("`%s` %s", f.Field, strVal))
		}
	}

	if len(query.ancestors) > 0 {
		for _, key := range query.ancestors {
			keyStr := fmt.Sprintf("%%%s", key.String())
			keyStr2 := fmt.Sprintf("%%%s/%%", key.String())
			where = append(where, fmt.Sprintf(
				"(`%s` LIKE %q OR `%s` LIKE %q)",
				FieldNameParent, keyStr, FieldNameParent, keyStr2))
		}
	}

	order := make([]string, 0)
	if len(query.orders) > 0 {
		for _, field := range query.orders {
			direction := "ASC"
			if field[:1] == "-" {
				field = field[1:]
				direction = "DESC"
			}
			order = append(order, fmt.Sprintf("`%s` %s", field, direction))
		}
	}

	order = append(order, fmt.Sprintf("CONCAT(`%s`,%q,`%s`) ASC", FieldNameParent, "/", FieldNameKey))

	locked := ""
	if x.mode == modeTransaction {
		switch query.lockMode {
		case lockForShare:
			locked = "LOCK IN SHARE MODE"

		case lockForUpdate:
			locked = "FOR UPDATE"

		default:
		}
	}

	stmt := &Statement{
		Where:  where,
		Order:  order,
		Limit:  query.limit,
		Locked: locked,
	}

	return stmt, nil
}

func (x *SQLAdapter) appendStatement(e *Entity, q *Query) *Query {
	if e.SoftDelete != nil {
		if !q.hasTrashed {
			q.filters = append(q.filters,
				newFilter(FieldNameSoftDelete, "=", nil, operatorMappingList["="]))
		}
	}
	return q
}

// toSQLSchema :
func (x *SQLAdapter) toSQLSchema(cols []*Field) []string {
	scripts := make([]string, 0)

	for _, item := range cols {
		schema := item.Schema
		props := make([]string, 0)

		if schema.IsUnsigned {
			props = append(props, "UNSIGNED")
		}

		// Set character set
		if schema.CharSet != nil {
			props = append(props,
				fmt.Sprintf("CHARACTER SET `%s` COLLATE `%s`",
					schema.CharSet.Encoding,
					schema.CharSet.Collation))
		}

		// Set not null if not nullable
		if !schema.IsNullable {
			props = append(props, "NOT NULL")

			strDefault := ""
			if schema.DefaultValue != nil {
				switch vt := schema.DefaultValue.(type) {
				case time.Time:
					strDefault = vt.Format("2006-01-02 15:04:05")
					strDefault = fmt.Sprintf("%q", strDefault)

				case datastore.GeoPoint:
					b, _ := json.Marshal(vt)
					strDefault = fmt.Sprintf("%q", string(b))

				case int, int8, int16, int32, int64:
					strDefault = fmt.Sprintf("%d", vt)

				case float32, float64:
					strDefault = fmt.Sprintf("%d", vt)

				case string:
					strDefault = fmt.Sprintf("%q", vt)

				case bool:
					strDefault = fmt.Sprintf("%t", vt)

				default:
				}
			}

			if strDefault != "" {
				props = append(props, fmt.Sprintf("DEFAULT %s", strDefault))
			}
		}

		if schema.IsUnique {
			props = append(props, "UNIQUE")
		}

		sql := fmt.Sprintf("`%s` %s", item.Name, schema.DataType)
		if len(props) > 0 {
			sql += " " + strings.Join(props, " ")
		}
		scripts = append(scripts, sql)
	}

	return scripts
}

// Close :
func (x *SQLAdapter) Close() error {
	return x.client.Close()
}
