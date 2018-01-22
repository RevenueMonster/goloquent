package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/datastore"
)

var operators = map[string]*operatorMapper{
	"=":    &operatorMapper{dbHybrid, eqOperatorToString},
	">":    &operatorMapper{dbHybrid, compareOperatorToString},
	"<":    &operatorMapper{dbHybrid, compareOperatorToString},
	">=":   &operatorMapper{dbHybrid, compareOperatorToString},
	"<=":   &operatorMapper{dbHybrid, compareOperatorToString},
	"IN":   &operatorMapper{dbHybrid, inOperatorToString},
	"!=":   &operatorMapper{dbMySQL, eqOperatorToString},
	"LIKE": &operatorMapper{dbMySQL, compareOperatorToString},
}

// Filter :
type Filter struct {
	Field      string
	Operator   string
	Value      interface{}
	stringFunc func(interface{}) (*string, error)
}

func (f *Filter) String() (*string, error) {
	return f.stringFunc(f.Value)
}

func newFilter(f string, o string, v interface{}, om *operatorMapper) *Filter {
	if v == nil {
		return &Filter{
			Field:      f,
			Operator:   o,
			Value:      v,
			stringFunc: om.StringFunc,
		}
	}

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return &Filter{
		Field:      f,
		Operator:   o,
		Value:      v,
		stringFunc: om.StringFunc,
	}
}

// Query :
type Query struct {
	table     *Table
	ancestors []*datastore.Key
	filters   []*Filter
	orders    []string
	limit     uint
	offset    uint
	errs      []error
}

func newQuery(t *Table) *Query {
	return &Query{
		table:     t,
		ancestors: make([]*datastore.Key, 0),
		filters:   make([]*Filter, 0),
		orders:    make([]string, 0),
		limit:     uint(0),
		offset:    uint(0),
		errs:      make([]error, 0),
	}
}

// First :
func (q *Query) First(modelStruct interface{}) error {
	return newBuilder(q).First(modelStruct)
}

// Get :
func (q *Query) Get(modelStruct interface{}) error {
	return newBuilder(q).Get(modelStruct)
}

// Paginate :
func (q *Query) Paginate(p *Pagination, modelStruct interface{}) error {
	return newBuilder(q).Paginate(p, modelStruct)
}

// Ancestor :
func (q *Query) Ancestor(ancestorKey *datastore.Key) *Query {
	table := q.table.name
	if ancestorKey.Incomplete() {
		q.errs = append(q.errs, errors.New("goloquent: invalid ancestor key (incomplete)"))
	}
	if ancestorKey.Name == table {
		q.errs = append(q.errs, errors.New("goloquent: cannot use current key as ancestor key"))
	}
	q.ancestors = append(q.ancestors, ancestorKey)
	return q
}

// Where :
func (q *Query) Where(field string, value interface{}) *Query {
	field = strings.TrimSpace(field)
	f := strings.Split(field, " ")
	if len(f) != 2 {
		panic(fmt.Errorf("goloquent: invalid field and operator"))
	}
	o := strings.TrimSpace(strings.ToUpper(f[1]))
	m, isOK := operators[o]
	if !isOK {
		panic(fmt.Errorf("goloquent: unsupported operator %v", o))
	}
	if m.Compatible != dbHybrid && m.Compatible != q.table.connection.db {
		panic(fmt.Errorf("goloquent: unsupported operator %v", o))
	}

	field = strings.TrimSpace(f[0])
	q.filters = append(q.filters, newFilter(field, o, value, m))
	return q
}

// Order :
func (q *Query) Order(fields interface{}) *Query {
	t := reflect.ValueOf(fields)

	switch t.Kind() {
	case reflect.String:
		q.orders = append(q.orders, strings.TrimSpace(fields.(string)))

	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() != reflect.String {
			panic(errors.New("goloquent: invalid order datatype"))
		}
		q.orders = append(q.orders, fields.([]string)...)

	default:
		panic(errors.New("goloquent: invalid order datatype"))

	}

	return q
}

// Update :
func (q *Query) Update(values interface{}) error {
	return newBuilder(q).UpdateMulti(values)
}

// Limit :
func (q *Query) Limit(i int) *Limit {
	return newLimit(q)
}
