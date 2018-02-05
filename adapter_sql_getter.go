package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
)

// Find :
func (x *SQLAdapter) Find(query *Query, key *datastore.Key, modelStruct interface{}) error {
	table := query.table.name
	var entity *Entity
	t := reflect.TypeOf(modelStruct)
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	query = x.appendStatement(entity, query)
	var stmt *Statement
	stmt, err = x.CompileStatement(query)
	if err != nil {
		return err
	}

	cond := fmt.Sprintf(
		"`%s` = %q AND `%s` = %q",
		FieldNameKey, stringPrimaryKey(key), FieldNameParent, key.Parent.String())
	q := fmt.Sprintf("SELECT * FROM `%s` WHERE %s", table, cond)
	if len(stmt.Where) > 0 {
		q += fmt.Sprintf(" AND %s", strings.Join(stmt.Where, " AND "))
	}
	q += " LIMIT 1"
	if len(stmt.Locked) > 0 {
		q += " " + stmt.Locked
	}
	q += ";"

	go x.sqlDebug(q)

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(q)
	if err != nil {
		return err
	}

	if len(results) <= 0 {
		return ErrNoSuchEntity
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

	t := reflect.TypeOf(modelStruct)
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	query = x.appendStatement(entity, query)
	var stmt *Statement
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
	q += " LIMIT 1"
	if len(stmt.Locked) > 0 {
		q += " " + stmt.Locked
	}
	q += ";"

	go x.sqlDebug(q)

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(q)
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

	query = x.appendStatement(entity, query)
	table := query.table.name
	stmt, err = x.CompileStatement(query)
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
	if stmt.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", stmt.Limit)
	}
	if len(stmt.Locked) > 0 {
		sql += " " + stmt.Locked
	}
	sql += ";"

	go x.sqlDebug(sql)

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

	return nil
}

// Paginate :
func (x *SQLAdapter) Paginate(query *Query, p *Pagination, modelStruct interface{}) error {
	table := query.table.name

	var entity *Entity
	t := reflect.TypeOf(modelStruct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	t = t.Elem()

	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	query = x.appendStatement(entity, query)
	var stmt *Statement
	stmt, err = x.CompileStatement(query)
	if err != nil {
		return err
	}

	sql := ""
	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if len(stmt.Order) > 0 {
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(stmt.Order, ","))
	}

	if p.Cursor != "" {
		var cursorKey *datastore.Key
		cursorKey, err = datastore.DecodeKey(p.Cursor)
		if err != nil {
			return errors.New("goloquent: invalid cursor key")
		}

		colKey := "RowNumber"
		sql2 := fmt.Sprintf(
			"%s JOIN (SELECT @ROW_NUM := 0) r",
			fmt.Sprintf(
				"SELECT @ROW_NUM := @ROW_NUM + 1 as RowNumber, `%s`, `%s` FROM `%s`",
				FieldNameParent, FieldNameKey, table)) + sql

		sql2 = fmt.Sprintf(
			"SELECT %s FROM (%s) AS Temp WHERE CONCAT(Temp.`%s`,%q,Temp.`%s`) = %q;",
			colKey, sql2, FieldNameParent, "/", FieldNameKey,
			cursorKey.Parent.String()+"/"+stringPrimaryKey(cursorKey))

		resp := make([]map[string][]byte, 0)
		resp, err = x.ExecQuery(sql2)
		if err != nil {
			return err
		}

		var offset int64
		offset, err = strconv.ParseInt(string(resp[0][colKey]), 10, 64)
		if err != nil {
			return err
		}

		if offset > 0 {
			offset--
		}
		query.offset = uint(offset)

		go x.sqlDebug(sql2)
	}

	sql = fmt.Sprintf("SELECT * FROM `%s`", table) + sql

	cap := p.Limit
	if cap <= 0 {
		cap = DefaultTotalRecord
	} else if cap > MaxRecord {
		cap = MaxRecord
	}
	cap++ // extra one record for pagination
	sql += fmt.Sprintf(" LIMIT %d", cap)

	if query.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", query.offset)
	}

	var total uint
	total, err = x.Count(query)
	if err != nil {
		return err
	}

	go x.sqlDebug(sql)

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

	copy := reflect.MakeSlice(slice.Type(), slice.Len(), slice.Len())
	if slice.Len() > 0 {
		count := uint(slice.Len())
		if slice.Len() > int(p.Limit) && entity.PrimaryKey != nil {
			// Get last record
			count--
			last := slice.Index(int(count))
			r := reflect.Indirect(last)
			pk := r.FieldByIndex(entity.PrimaryKey.Index)
			if pk.IsValid() {
				p.Cursor = pk.Interface().(*datastore.Key).Encode()
			}
			copy = reflect.MakeSlice(slice.Type(), int(count), int(count))
		} else {
			p.Cursor = ""
		}
		p.Count = count
	}

	reflect.Copy(copy, slice)

	// Sync pagination data
	p.Total = total

	iv := reflect.Indirect(reflect.ValueOf(modelStruct))
	iv.Set(copy)

	return nil
}
