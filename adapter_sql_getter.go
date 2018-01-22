package goloquent

import (
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/fatih/color"
)

// Find :
func (x *SQLAdapter) Find(query *Query, key *datastore.Key, modelStruct interface{}) error {
	table := query.table.name
	cond := fmt.Sprintf(
		"`%s` = %q AND `%s` = %q",
		FieldNameKey, stringPrimaryKey(key), FieldNameParent, key.Parent.String())

	q := fmt.Sprintf("SELECT * FROM `%s` WHERE %s LIMIT 1;", table, cond)
	fmt.Println("************* START FIND QUERY ************")
	fmt.Println(color.GreenString(q))
	fmt.Println("************* ENDED FIND QUERY ************")

	// results := make([]map[string][]byte, 0)
	results, err := x.ExecQuery(q)
	if err != nil {
		return err
	}

	if len(results) <= 0 {
		return ErrNoSuchEntity
	}

	var entity *Entity
	t := reflect.TypeOf(modelStruct)
	entity, err = getEntity(t)
	if err != nil {
		return err
	}

	var slice reflect.Value
	slice, err = x.mapResults(query, entity, t, results)
	if err != nil {
		return err
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	o := reflect.Indirect(slice.Index(0))
	v.Set(o)

	return nil
}

// First :
func (x *SQLAdapter) First(query *Query, modelStruct interface{}) error {
	table := query.table.name
	stmt, err := x.CompileStatement(query)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT * FROM `%s`", table)
	if len(stmt.Where) > 0 {
		q += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if len(stmt.Order) > 0 {
		q += fmt.Sprintf(" ORDER BY %s", strings.Join(stmt.Order, ","))
	}
	q += " LIMIT 1;"

	fmt.Println("************* START FIRST QUERY ************")
	fmt.Println(color.GreenString(q))
	fmt.Println("************* ENDED FIRST QUERY ************")

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(q)
	if err != nil {
		return err
	}

	if len(results) <= 0 {
		return nil
	}

	var entity *Entity
	t := reflect.TypeOf(modelStruct)
	entity, err = getEntity(t)
	if err != nil {
		return err
	}

	var slice reflect.Value
	slice, err = x.mapResults(query, entity, t, results)
	if err != nil {
		return err
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	o := reflect.Indirect(slice.Index(0))
	v.Set(o)

	return nil
}

// Get :
func (x *SQLAdapter) Get(query *Query, modelStruct interface{}) error {
	var (
		stmt *Statement
		err  error
	)

	table := query.table.name
	stmt, err = x.CompileStatement(query)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT * FROM `%s`", table)
	if len(stmt.Where) > 0 {
		q += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if len(stmt.Order) > 0 {
		q += fmt.Sprintf(" ORDER BY %s", strings.Join(stmt.Order, ","))
	}
	if stmt.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", stmt.Limit)
	}
	q += ";"

	fmt.Println("************* START GET QUERY ************")
	fmt.Println(color.GreenString(q))
	fmt.Println("************* ENDED GET QUERY ************")

	// cols := entity.GetFields()
	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(q)
	if err != nil {
		return err
	}

	if len(results) <= 0 {
		return nil
	}

	var entity *Entity
	t := reflect.TypeOf(modelStruct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	t = t.Elem()

	entity, err = getEntity(t)
	if err != nil {
		return err
	}

	var slice reflect.Value
	slice, err = x.mapResults(query, entity, t, results)
	if err != nil {
		return err
	}

	iv := reflect.Indirect(reflect.ValueOf(modelStruct))
	iv.Set(slice)

	return nil
}

// Paginate :
func (x *SQLAdapter) Paginate(query *Query, p *Pagination, modelStruct interface{}) error {
	table := query.table.name

	stmt, err := x.CompileStatement(query)
	if err != nil {
		return err
	}

	var entity *Entity
	t := reflect.TypeOf(modelStruct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	t = t.Elem()

	entity, err = getEntity(t)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("SELECT * FROM `%s`", table)
	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if len(stmt.Order) > 0 {
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(stmt.Order, ","))
	}
	cap := p.Limit
	if cap <= 0 || cap > MaxRecord {
		cap = MaxRecord
	}
	sql += fmt.Sprintf(" LIMIT %d", cap)

	var total int
	total, err = x.Count(query)
	if err != nil {
		return err
	}

	fmt.Println("************* START PAGINATE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED PAGINATE QUERY ************")

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(sql)
	if err != nil {
		return err
	}

	if len(results) <= 0 {
		return nil
	}

	var slice reflect.Value
	slice, err = x.mapResults(query, entity, t, results)
	if err != nil {
		return err
	}

	iv := reflect.Indirect(reflect.ValueOf(modelStruct))
	iv.Set(slice)

	// Get last record
	last := slice.Index(slice.Len() - 1)
	fmt.Println(last)

	// Sync pagination data
	p.Total = uint(total)
	p.Count = uint(len(results))
	p.Cursor = ""

	return nil
}
