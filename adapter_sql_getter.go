package goloquent

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
)

// Find :
func (x *SQLAdapter) Find(query *Query, key *datastore.Key, modelStruct interface{}) error {
	var entity *Entity
	t := reflect.TypeOf(modelStruct)
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	if len(query.tables) < 1 {
		query.tables = append(query.tables, entity.name)
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
	q := fmt.Sprintf("SELECT * FROM (%s) AS Master WHERE %s", strings.Join(stmt.Tables, " UNION ALL "), cond)
	if len(stmt.Where) > 0 {
		q += fmt.Sprintf(" AND %s", strings.Join(stmt.Where, " AND "))
	}
	q += " LIMIT 1"
	if len(stmt.Locked) > 0 {
		q += " " + stmt.Locked
	}
	q += ";"

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
	t := reflect.TypeOf(modelStruct)
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	if len(query.tables) < 1 {
		query.tables = append(query.tables, entity.name)
	}

	query = x.appendStatement(entity, query)
	var stmt *Statement
	stmt, err = x.CompileStatement(query)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT * FROM (%s) AS Master",
		strings.Join(stmt.Tables, " UNION ALL "))
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

type Cursor struct {
	cc []byte
}

// String returns a base-64 string representation of a cursor.
func (c Cursor) String() string {
	if c.cc == nil {
		return ""
	}

	return strings.TrimRight(base64.URLEncoding.EncodeToString(c.cc), "=")
}

func (c Cursor) Offset() int64 {
	v, _ := strconv.ParseInt(string(c.cc), 10, 64)
	return v
}

// Decode decodes a cursor from its base-64 string representation.
func DecodeCursor(s string) (Cursor, error) {
	if s == "" {
		return Cursor{}, nil
	}
	if n := len(s) % 4; n != 0 {
		s += strings.Repeat("=", 4-n)
	}
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, err
	}
	return Cursor{b}, nil
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

	if len(query.tables) < 1 {
		query.tables = append(query.tables, entity.name)
	}

	query = x.appendStatement(entity, query)
	stmt, err = x.CompileStatement(query)
	if err != nil {
		return err
	}

	sql := stmt.Tables[0]
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

	if len(query.tables) < 1 {
		query.tables = append(query.tables, entity.name)
	}

	query = x.appendStatement(entity, query)
	var stmt *Statement
	stmt, err = x.CompileStatement(query)
	if err != nil {
		return err
	}

	// if p.Cursor != "" {

	// 	colKey := "RowNumber"
	// 	sql2 := fmt.Sprintf(
	// 		"%s JOIN (SELECT @ROW_NUM := 0) AS Record",
	// 		fmt.Sprintf(
	// 			"SELECT @ROW_NUM := @ROW_NUM + 1 as RowNumber, `%s`, `%s` FROM (%s) AS Master",
	// 			FieldNameParent, FieldNameKey, strings.Join(stmt.Table, " UNION ALL "))) + sql

	// 	sql2 = fmt.Sprintf(
	// 		"SELECT %s FROM (%s) AS Temp WHERE CONCAT(Temp.`%s`,%q,Temp.`%s`) = %q;",
	// 		colKey, sql2, FieldNameParent, "/", FieldNameKey,
	// 		cursorKey.Parent.String()+"/"+stringPrimaryKey(cursorKey))

	// 	resp := make([]map[string][]byte, 0)
	// 	resp, err = x.ExecQuery(sql2)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	var offset int64
	// 	offset, err = strconv.ParseInt(string(resp[0][colKey]), 10, 64)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if offset > 0 {
	// 		offset--
	// 	}
	// 	query.offset = uint(offset)
	// }

	offset := int64(0)
	args := make([]interface{}, 0)
	if p.Cursor != "" {
		cursorKey, err := DecodeCursor(p.Cursor)
		if err != nil {
			return errors.New("goloquent: invalid cursor key")
		}
		offset = cursorKey.Offset()
	}
	stmt.Order = append(stmt.Order, "`$PrimaryKey` ASC")

	sql := ""
	selectStmt := stmt.Tables[0]
	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if len(stmt.Order) > 0 {
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(stmt.Order, ","))
	}
	sql = selectStmt + sql

	cap := p.Limit
	if cap <= 0 {
		cap = DefaultTotalRecord
	} else if cap > MaxRecordGet {
		cap = MaxRecordGet
	}
	cap++ // extra one record for pagination
	sql += fmt.Sprintf(" LIMIT %d", cap)

	if offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", offset)
	}

	// var total uint
	// total, err = x.Count(query)
	// if err != nil {
	// 	return err
	// }

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(sql, args...)
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
		cc := Cursor{cc: []byte(fmt.Sprintf("%d", p.Limit))}
		p.Cursor = cc.String()
		if slice.Len() > int(p.Limit) && entity.PrimaryKey != nil {
			count--
			offset = offset + int64(p.Limit)
			p.Cursor = (Cursor{[]byte(fmt.Sprintf("%d", offset))}).String()
			// Get last record
			// last := slice.Index(int(count))
			// r := reflect.Indirect(last)
			// pk := r.FieldByIndex(entity.PrimaryKey.Index)
			// if pk.IsValid() {
			// 	p.Cursor = pk.Interface().(*datastore.Key).Encode()
			// }
			copy = reflect.MakeSlice(slice.Type(), int(count), int(count))
		} else {
			p.Cursor = ""
		}
		p.Count = count
	}

	reflect.Copy(copy, slice)

	// Sync pagination data
	// p.Total = total

	iv := reflect.Indirect(reflect.ValueOf(modelStruct))
	iv.Set(copy)

	return nil
}
