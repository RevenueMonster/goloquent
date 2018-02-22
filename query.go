package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/datastore"
)

var operatorMappingList = map[string]*operatorMapper{
	"=":        &operatorMapper{dbHybrid, eqOperatorToString},
	">":        &operatorMapper{dbHybrid, compareOperatorToString},
	"<":        &operatorMapper{dbHybrid, compareOperatorToString},
	">=":       &operatorMapper{dbHybrid, compareOperatorToString},
	"<=":       &operatorMapper{dbHybrid, compareOperatorToString},
	"IN":       &operatorMapper{dbHybrid, inOperatorToString},
	"!=":       &operatorMapper{dbMySQL, eqOperatorToString},
	"LIKE":     &operatorMapper{dbMySQL, compareOperatorToString},
	"NOT LIKE": &operatorMapper{dbMySQL, compareOperatorToString},
}

// Filter :
type Filter struct {
	Field      string
	Operator   string
	Value      interface{}
	stringFunc func(interface{}) (*string, error)
}

func (f Filter) String() (*string, error) {
	return f.stringFunc(f.Value)
}

func newFilter(f string, o string, v interface{}, om *operatorMapper) Filter {
	if v == nil {
		return Filter{
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

	return Filter{
		Field:      f,
		Operator:   o,
		Value:      v,
		stringFunc: om.StringFunc,
	}
}

// Query :
type Query struct {
	table      *Table
	tables     []string
	ancestors  []*datastore.Key
	filters    []Filter
	orders     []string
	lockMode   string
	limit      uint
	offset     uint
	errs       []error
	hasTrashed bool
}

func newQuery(t *Table) *Query {
	return &Query{
		table:     t,
		tables:    []string{t.name},
		ancestors: make([]*datastore.Key, 0),
		filters:   make([]Filter, 0),
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

// Find :
func (q *Query) Find(key *datastore.Key, modelStruct interface{}) error {
	return newBuilder(q).Find(key, modelStruct)
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

// LockForShared :
func (q *Query) LockForShared() *Getter {
	q.lockMode = lockForShare
	return newGetter(q)
}

// LockForUpdate :
func (q *Query) LockForUpdate() *Getter {
	q.lockMode = lockForUpdate
	return newGetter(q)
}

// Where :
func (q *Query) Where(field interface{}, operator string, value interface{}) *Query {
	strField := ""
	switch field.(type) {
	case string:
		strField = field.(string)

	case RawQuery:
		rq := field.(*RawQuery)
		strField = rq.Name

	default:
		q.errs = append(q.errs, errors.New("goloquent: invalid field datatype"))
		return q
	}

	strField = strings.TrimSpace(strField)
	operator = strings.TrimSpace(strings.ToUpper(operator))
	m, isOK := operatorMappingList[operator]
	if !isOK {
		panic(fmt.Errorf("goloquent: unsupported operator %v", operator))
	}
	if m.Compatible != dbHybrid && m.Compatible != q.table.connection.db {
		panic(fmt.Errorf("goloquent: unsupported operator %v", operator))
	}

	if value != nil {
		v := reflect.ValueOf(value)
		if operator != "IN" && v.Type() != typeOfByte &&
			(v.Kind() == reflect.Array || v.Kind() == reflect.Slice) {
			f := make([]Filter, v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				f[i] = newFilter(strField, operator, v.Index(i).Interface(), m)
			}
			q.filters = append(q.filters, f...)
			return q
		}
	}

	q.filters = append(q.filters, newFilter(strField, operator, value, m))

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

// Count :
func (q *Query) Count() (uint, error) {
	return newBuilder(q).Count()
}

// Sum :
func (q *Query) Sum(field string) (int, error) {
	adapter, err := q.table.getSQLAdapter()
	if err != nil {
		return 0, err
	}
	return adapter.Sum(field, q)
}

// WithTrashed :
func (q *Query) WithTrashed() *Query {
	q.hasTrashed = true
	return q
}

// Update :
func (q *Query) Update(values interface{}) error {
	return newBuilder(q).UpdateMulti(values)
}

// Delete :
func (q *Query) Delete() error {
	adapter, err := q.table.getSQLAdapter()
	if err != nil {
		return err
	}
	return adapter.deleteWithQuery(q)
}

// Limit :
func (q *Query) Limit(i int) *Limit {
	q.limit = uint(i)
	return newLimit(q)
}
