package goloquent

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"

	"cloud.google.com/go/datastore"
)

// Create :
func (x *SQLAdapter) Create(query *Query, modelStruct interface{}, parentKey interface{}) error {
	t := reflect.TypeOf(modelStruct).Elem()

	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	table := query.table.name
	cols := entity.GetFields()

	// Generate primary key before insert to database
	primaryKey := x.newPrimaryKey(table, nil)
	if parentKey != nil {
		primaryKey = x.newPrimaryKey(table, parentKey.(*datastore.Key))
	}

	fields := make([]string, 0)
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameKey))
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameParent))

	vals := make([]string, 0)
	vals = append(vals, fmt.Sprintf("%q", stringPrimaryKey(primaryKey)))
	vals = append(vals, fmt.Sprintf("%q", primaryKey.Parent.String()))

	// Call datastore.PropertyLoadSaver's Save func
	// _, err = entity.SaveFunc(v.Interface())
	// if err != nil {
	// 	return err
	// }

	// Run through every property in struct and convert to string
	v := reflect.ValueOf(modelStruct)
	for _, fs := range cols {
		f := v.Elem().FieldByIndex(fs.Index)
		if !f.IsValid() {
			return fmt.Errorf("goloquent: missing field on index %v", fs.Index)
		}

		// Skip primary key
		if fs.IsPrimaryKey {
			f.Set(reflect.ValueOf(primaryKey))
			continue
		}

		str, err := fs.String(f.Interface())
		if err != nil {
			return err
		}

		val := "NULL"
		if str != nil {
			val = fmt.Sprintf("%s", *str)
			if fs.Schema.IsEscape {
				val = fmt.Sprintf("%q", *str)
			}
		}

		fields = append(fields, fmt.Sprintf("`%s`", fs.Name))
		vals = append(vals, fmt.Sprintf("%s", val))
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%v) VALUES (%v);",
		table, strings.Join(fields, ","), strings.Join(vals, ","))

	fmt.Println("************* START CREATE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED CREATE QUERY ************")

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// CreateMulti :
func (x *SQLAdapter) CreateMulti(query *Query, modelStruct interface{}, parentKey interface{}) error {
	t := reflect.TypeOf(modelStruct).Elem().Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.ValueOf(modelStruct).Elem()
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	table := query.table.name
	cols := entity.GetFields()

	pKeys := make([]*datastore.Key, v.Len())
	kv := reflect.Indirect(reflect.ValueOf(parentKey))
	for i := 0; i < v.Len(); i++ {
		var k *datastore.Key
		// TODO: allow nil in parent key
		if kv.Kind() == reflect.Slice {
			parent := kv.Index(i).Interface().(*datastore.Key)
			k = x.newPrimaryKey(table, parent)
		} else {
			k = x.newPrimaryKey(table, parentKey.(*datastore.Key))
		}
		pKeys[i] = k
	}

	fields := make([]string, 0)
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameKey))
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameParent))

	records := make([]string, 0)
	for i := 0; i < v.Len(); i++ {
		fv := v.Index(i)
		if !fv.IsValid() {
			return errors.New("goloquent (create multiple): invalid model data")
		}

		// Generate primary key before insert to database
		primaryKey := pKeys[i]
		strKey := stringPrimaryKey(primaryKey)

		vals := make([]string, 0)
		vals = append(vals, fmt.Sprintf("%q", strKey))
		vals = append(vals, fmt.Sprintf("%q", primaryKey.Parent.String()))

		// Call datastore.PropertyLoadSaver's Save func
		_, err = entity.SaveFunc(fv.Interface())
		if err != nil {
			return err
		}

		// Run through every property in struct and convert to string
		for _, fs := range cols {
			// Skip primary key
			if fs.IsPrimaryKey {
				continue
			}
			f := fv.Elem().FieldByIndex(fs.Index)
			if !f.IsValid() {
				return fmt.Errorf("goloquent: missing field %v", fs.Name)
			}

			str, err := fs.String(f.Interface())
			if err != nil {
				return err
			}

			val := "NULL"
			if str != nil {
				if fs.Schema.IsEscape {
					val = fmt.Sprintf("%q", *str)
				} else {
					val = fmt.Sprintf("%s", *str)
				}
			}

			if i == 0 {
				fields = append(fields, fmt.Sprintf("`%s`", fs.Name))
			}
			vals = append(vals, fmt.Sprintf("%s", val))
		}

		records = append(records, fmt.Sprintf("(%s)", strings.Join(vals, ",")))
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES %s;",
		table, strings.Join(fields, ","), strings.Join(records, ","))

	fmt.Println("************* START CREATE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED CREATE QUERY ************")

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// UpsertMulti :
func (x *SQLAdapter) UpsertMulti(query *Query, modelStruct interface{}, parentKey interface{}) error {
	t := reflect.TypeOf(modelStruct).Elem().Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	v := reflect.ValueOf(modelStruct).Elem()
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	table := query.table.name
	cols := entity.GetFields()

	pKeys := make([]*datastore.Key, v.Len())
	kv := reflect.Indirect(reflect.ValueOf(parentKey))
	for i := 0; i < v.Len(); i++ {
		var k *datastore.Key
		if kv.Kind() == reflect.Slice {
			parent := kv.Index(i).Interface().(*datastore.Key)
			k = x.newPrimaryKey(table, parent)
		} else {
			k = x.newPrimaryKey(table, parentKey.(*datastore.Key))
		}
		pKeys[i] = k
	}

	fields := make([]string, 0)
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameKey))
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameParent))

	colNames := make([]string, 0)

	records := make([]string, 0)
	for i := 0; i < v.Len(); i++ {
		fv := v.Index(i)
		if !fv.IsValid() {
			return errors.New("goloquent (upsert multiple): invalid model data")
		}

		// Generate primary key before insert to database
		primaryKey := pKeys[i]
		strKey := stringPrimaryKey(primaryKey)

		vals := make([]string, 0)
		vals = append(vals, fmt.Sprintf("%q", strKey))
		vals = append(vals, fmt.Sprintf("%q", primaryKey.Parent.String()))

		// Call datastore.PropertyLoadSaver's Save func
		_, err = entity.SaveFunc(fv.Interface())
		if err != nil {
			return err
		}

		// Run through every property in struct and convert to string
		for _, fs := range cols {
			// Skip primary key
			if fs.IsPrimaryKey {
				continue
			}
			f := fv.Elem().FieldByIndex(fs.Index)
			if !f.IsValid() {
				return fmt.Errorf("goloquent: missing field %v", fs.Name)
			}

			str, err := fs.String(f.Interface())
			if err != nil {
				return err
			}

			val := "NULL"
			if str != nil {
				val = fmt.Sprintf("%s", *str)
				if fs.Schema.IsEscape {
					val = fmt.Sprintf("%q", *str)
				}
			}

			if i == 0 {
				name := fs.Name
				colNames = append(colNames, fmt.Sprintf("`%s`=VALUES(`%s`)", name, name))
				fields = append(fields, fmt.Sprintf("`%s`", name))
			}
			vals = append(vals, fmt.Sprintf("%s", val))
		}

		records = append(records, fmt.Sprintf("(%s)", strings.Join(vals, ",")))
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES %s ON DUPLICATE KEY UPDATE %s;",
		table, strings.Join(fields, ","), strings.Join(records, ","), strings.Join(colNames, ","))

	fmt.Println("************* START CREATE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED CREATE QUERY ************")

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// Update :
func (x *SQLAdapter) Update(query *Query, modelStruct interface{}) error {
	// v := reflect.ValueOf(modelStruct).Elem()
	// t := reflect.TypeOf(modelStruct).Elem()

	// columns, err := getColumns(t)
	// if err != nil {
	// 	return err
	// }

	// primaryKey := datastore.IncompleteKey(x.table, nil)
	// values := make([]string, 0)
	// for _, s := range columns {
	// 	f := v.FieldByName(s.Name)
	// 	if s.IsPrimaryKey {
	// 		primaryKey = f.Interface().(*datastore.Key)
	// 		continue
	// 	}
	// 	strValue := ""
	// 	values = append(values, fmt.Sprintf("`%s` = %s", s.FieldName, strValue))
	// }

	// if primaryKey.Incomplete() {
	// 	return errors.New("invalid model struct (primary key not found)")
	// }

	// strKey := fmt.Sprintf("%q", strconv.FormatInt(primaryKey.ID, 10))
	// if primaryKey.Parent != nil {
	// 	strKey = primaryKey.Parent.String()
	// }

	// condition := fmt.Sprintf("`%s` = %s AND `%s` = %q", FieldNameKey, strKey, FieldNameParent, primaryKey.Parent.String())
	// sql := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", query.table.name, strings.Join(values, ","), condition)

	// fmt.Println("************* START UPDATE QUERY ************")
	// fmt.Println(sql)
	// fmt.Println("************* ENDED UPDATE QUERY ************")

	// if _, err := a.client.Query(sql); err != nil {
	// 	return err
	// }

	return nil
}

// UpdateMulti :
func (x *SQLAdapter) UpdateMulti(query *Query, modelStruct interface{}) error {

	table := query.table.name
	stmt, err := x.CompileStatement(query)
	if err != nil {
		return err
	}

	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	vals := make([]string, 0)
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			f := v.MapIndex(key)
			t := reflect.TypeOf(f.Interface())
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}

			parseFunc, isValid := interfaceToStringList[t]
			if !isValid {
				return ErrUnsupportDataType
			}

			val := "NULL"
			str, _ := parseFunc(f.Interface())
			if str != nil {
				val = fmt.Sprintf("%s", *str)
				if !isNumber(t) {
					val = fmt.Sprintf("%q", *str)
				}
			}

			vals = append(vals, fmt.Sprintf("`%s` = %s", key, val))
		}
	} else {
		// TODO: struct
	}

	if len(vals) <= 0 {
		return errors.New("goloquent: no update field provided")
	}

	sql := fmt.Sprintf("UPDATE `%s` SET %s", table, strings.Join(vals, ","))
	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, ","))
	}
	sql += ";"

	fmt.Println("************* START UPDATE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED UPDATE QUERY ************")

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// Delete :
func (x *SQLAdapter) Delete(query *Query, key *datastore.Key) error {
	where := fmt.Sprintf(
		"WHERE `%s` = %q AND `%s` = %q",
		FieldNameKey, stringPrimaryKey(key), FieldNameParent, key.Parent.String())
	sql := fmt.Sprintf("DELETE FROM `%s` %s", query.table.name, where)
	fmt.Println("************* START DELETE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED DELETE QUERY ************")
	if _, err := x.Exec(sql); err != nil {
		return err
	}
	return nil
}

// DeleteMulti :
func (x *SQLAdapter) DeleteMulti(query *Query, keys []*datastore.Key) error {
	pks := make([]string, 0)
	for _, key := range keys {
		strPK := fmt.Sprintf("%q", stringPrimaryKey(key)+key.Parent.String())
		pks = append(pks, strPK)
	}

	table := query.table.name
	list := fmt.Sprintf(
		"SELECT CONCAT(`%s`, `%s`) AS `%s`",
		FieldNameKey,
		FieldNameParent,
		FieldNamePrimaryKey)
	sql := fmt.Sprintf(
		"DELETE FROM `%s` WHERE (%s) IN (%s)",
		table, list, strings.Join(pks, ","))

	fmt.Println("************* START DELETE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED DELETE QUERY ************")
	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// SoftDelete :
func (x *SQLAdapter) SoftDelete(query *Query, key *datastore.Key) error {
	value := fmt.Sprintf("`DeletedDateTime` = %q", time.Now().UTC().Format(MySQLDateTimeFormat))
	where := fmt.Sprintf("WHERE `%s` = %q AND `%s` = %q",
		FieldNameKey, stringPrimaryKey(key), FieldNameParent, key.Parent.String())
	sql := fmt.Sprintf("UPDATE `%s` SET %s %s", query.table.name, value, where)
	fmt.Println("************* START SOFT DELETE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED SOFT DELETE QUERY ************")
	if _, err := x.Exec(sql); err != nil {
		return err
	}
	return nil
}

// RunInTransaction :
func (x *SQLAdapter) RunInTransaction(table *Table, callback func(*Connection) error) error {
	txn, err := x.client.Begin()
	if err != nil {
		return err
	}
	c := &Connection{
		db: dbMySQL,
		adapter: &SQLAdapter{
			mode:   modeTransaction,
			client: x.client,
			txn:    txn,
		},
	}
	return func(c *Connection, txn *sql.Tx) error {
		if err := callback(c); err != nil {
			return txn.Rollback()
		}
		return txn.Commit()
	}(c, txn)
}
