package goloquent

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

func getTableName(entity *Entity, query *Query) string {
	if query.table.name != "" {
		return query.table.name
	}
	return entity.name
}

func getField(v reflect.Value, path []int) reflect.Value {
	for _, p := range path {
		v = v.Field(p)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return reflect.Zero(v.Type())
		}
	}
	return v
}

// Create :
func (x *SQLAdapter) Create(query *Query, modelStruct interface{}, parentKey *datastore.Key) error {
	t := reflect.TypeOf(modelStruct).Elem()

	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	table := getTableName(entity, query)
	cols := entity.GetFields()

	// Generate primary key before insert to database
	primaryKey := x.newPrimaryKey(table, parentKey)

	fields := make([]string, 0)
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameKey))
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameParent))

	vals := make([]string, 0)
	vals = append(vals, fmt.Sprintf("%q", stringPrimaryKey(primaryKey)))
	vals = append(vals, fmt.Sprintf("%q", primaryKey.Parent.String()))

	// Run through every property in struct and convert to string
	v := reflect.ValueOf(modelStruct)
	if entity.PrimaryKey != nil {
		f := v.Elem().FieldByIndex(entity.PrimaryKey.Index)
		f.Set(reflect.ValueOf(primaryKey))
	}

	// Call datastore.PropertyLoadSaver's Save func
	_, err = entity.SaveFunc(v.Interface())
	if err != nil {
		return err
	}

	for _, fs := range cols {
		f := getField(v.Elem(), fs.Index)
		if !f.IsValid() {
			return fmt.Errorf("goloquent: missing field on index %v", fs.Index)
		}

		str, err := fs.String(f.Interface())
		if err != nil {
			return err
		}

		val := "NULL"
		if str != nil {
			if !(isZero(*str) && fs.Schema.IsNullable) {
				val = fmt.Sprintf("%s", *str)
				if fs.Schema.IsEscape {
					val = fmt.Sprintf("%q", *str)
				}
			}
		}

		fields = append(fields, fmt.Sprintf("`%s`", fs.Name))
		vals = append(vals, fmt.Sprintf("%s", val))
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%v) VALUES (%v);",
		table, strings.Join(fields, ","), strings.Join(vals, ","))

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

	table := getTableName(entity, query)
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

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		// Run through every property in struct and convert to string
		for _, fs := range cols {
			f := getField(fv, fs.Index)
			if !f.IsValid() {
				return fmt.Errorf("goloquent: missing field %v", fs.Name)
			}

			str, err := fs.String(f.Interface())
			if err != nil {
				return err
			}

			val := "NULL"
			if str != nil {
				if !(isZero(*str) && fs.Schema.IsNullable) {
					val = fmt.Sprintf("%s", *str)
					if fs.Schema.IsEscape {
						val = fmt.Sprintf("%q", *str)
					}
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

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// Upsert :
func (x *SQLAdapter) Upsert(query *Query, modelStruct interface{}, parentKey *datastore.Key, excluded ...string) error {
	t := reflect.TypeOf(modelStruct).Elem()

	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	table := getTableName(entity, query)
	cols := entity.GetFields()

	excludedFields := make(map[string]bool, 0)
	for _, fieldName := range excluded {
		if fieldName == FieldNameKey ||
			fieldName == FieldNameParent ||
			fieldName == "__key__" {
			continue
		}
		excludedFields[fieldName] = true
	}

	// Generate primary key before insert to database
	primaryKey := x.newPrimaryKey(table, parentKey)

	fields := make([]string, 0)
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameKey))
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameParent))

	vals := make([]string, 0)
	vals = append(vals, fmt.Sprintf("%q", stringPrimaryKey(primaryKey)))
	vals = append(vals, fmt.Sprintf("%q", primaryKey.Parent.String()))

	// Run through every property in struct and convert to string
	where := make([]string, 0)
	v := reflect.ValueOf(modelStruct)
	if entity.PrimaryKey != nil {
		f := v.Elem().FieldByIndex(entity.PrimaryKey.Index)
		f.Set(reflect.ValueOf(primaryKey))
	}

	if entity.SoftDelete != nil {
		fields = append(fields, fmt.Sprintf("`%s`", FieldNameSoftDelete))
		f := v.Elem().FieldByIndex(entity.SoftDelete.Index)
		sd := f.Interface().(SoftDelete)
		strSoftDelete := "NULL"
		if !isZero(sd.DeletedDateTime) {
			strSoftDelete = fmt.Sprintf("%q", sd.DeletedDateTime.Format(MySQLDateTimeFormat))
		}
		where = append(where, fmt.Sprintf("`%s`=VALUES(`%s`)", FieldNameSoftDelete, FieldNameSoftDelete))
		vals = append(vals, strSoftDelete)
	}

	for _, fs := range cols {
		f := getField(v.Elem(), fs.Index)
		if !f.IsValid() {
			return fmt.Errorf("goloquent: missing field on index %v", fs.Index)
		}

		str, err := fs.String(f.Interface())
		if err != nil {
			return err
		}

		val := "NULL"
		if str != nil {
			if !(isZero(*str) && fs.Schema.IsNullable) {
				val = fmt.Sprintf("%s", *str)
				if fs.Schema.IsEscape {
					val = fmt.Sprintf("%q", *str)
				}
			}
		}

		if _, isOk := excludedFields[fs.Name]; !isOk {
			where = append(where, fmt.Sprintf("`%s`=VALUES(`%s`)", fs.Name, fs.Name))
		}

		fields = append(fields, fmt.Sprintf("`%s`", fs.Name))
		vals = append(vals, fmt.Sprintf("%s", val))
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%v) VALUES (%v) ON DUPLICATE KEY UPDATE %s;",
		table, strings.Join(fields, ","), strings.Join(vals, ","), strings.Join(where, ","))

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// UpsertMulti :
func (x *SQLAdapter) UpsertMulti(query *Query, modelStruct interface{}, parentKey interface{}, excluded ...string) error {
	t := reflect.TypeOf(modelStruct).Elem().Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(modelStruct).Elem()
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	table := getTableName(entity, query)
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

	excludedFields := make(map[string]bool, 0)
	for _, fieldName := range excluded {
		if fieldName == FieldNameKey ||
			fieldName == FieldNameParent ||
			fieldName == "__key__" {
			continue
		}
		excludedFields[fieldName] = true
	}

	fields := make([]string, 0)
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameKey))
	fields = append(fields, fmt.Sprintf("`%s`", FieldNameParent))

	colNames := make([]string, 0)

	records := make([]string, 0)
	for i := 0; i < v.Len(); i++ {
		fv := v.Index(i)
		if !fv.IsValid() {
			return errors.New("goloquent: invalid model data")
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

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		// Run through every property in struct and convert to string
		for _, fs := range cols {
			f := getField(fv, fs.Index)
			if !f.IsValid() {
				return fmt.Errorf("goloquent: missing field %v", fs.Name)
			}

			str, err := fs.String(f.Interface())
			if err != nil {
				return err
			}

			val := "NULL"
			if str != nil {
				if !(isZero(*str) && fs.Schema.IsNullable) {
					val = fmt.Sprintf("%s", *str)
					if fs.Schema.IsEscape {
						val = fmt.Sprintf("%q", *str)
					}
				}
			}

			if i == 0 {
				name := fs.Name
				if _, isOk := excludedFields[fs.Name]; !isOk {
					colNames = append(colNames, fmt.Sprintf("`%s`=VALUES(`%s`)", name, name))
				}
				fields = append(fields, fmt.Sprintf("`%s`", name))
			}
			vals = append(vals, fmt.Sprintf("%s", val))
		}

		if entity.SoftDelete != nil {
			if i == 0 {
				colNames = append(colNames, fmt.Sprintf("`%s`=VALUES(`%s`)", FieldNameSoftDelete, FieldNameSoftDelete))
				fields = append(fields, fmt.Sprintf("`%s`", FieldNameSoftDelete))
			}
			f := fv.FieldByIndex(entity.SoftDelete.Index)
			sd := f.Interface().(SoftDelete)
			strSoftDelete := "NULL"
			if !isZero(sd.DeletedDateTime) {
				strSoftDelete = fmt.Sprintf("%q", sd.DeletedDateTime.Format(MySQLDateTimeFormat))
			}
			vals = append(vals, strSoftDelete)
		}

		records = append(records, fmt.Sprintf("(%s)", strings.Join(vals, ",")))
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES %s ON DUPLICATE KEY UPDATE %s;",
		table, strings.Join(fields, ","), strings.Join(records, ","), strings.Join(colNames, ","))

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// Update :
func (x *SQLAdapter) Update(query *Query, modelStruct interface{}) error {
	t := reflect.TypeOf(modelStruct).Elem()

	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	table := getTableName(entity, query)
	cols := entity.GetFields()
	vals := make([]string, 0)

	// Run through every property in struct and convert to string
	v := reflect.ValueOf(modelStruct)
	if entity.PrimaryKey == nil {
		return ErrMissingPrimaryKey
	}

	// Call datastore.PropertyLoadSaver's Save func
	_, err = entity.SaveFunc(modelStruct)
	if err != nil {
		return err
	}

	k := v.Elem().FieldByIndex(entity.PrimaryKey.Index)
	if !k.IsValid() || k.IsNil() {
		return ErrMissingPrimaryKey
	}
	primaryKey := k.Interface().(*datastore.Key)

	for _, fs := range cols {
		f := getField(v.Elem(), fs.Index)
		if !f.IsValid() {
			return fmt.Errorf("goloquent: missing field on index %v", fs.Index)
		}

		var str *string
		var err error
		if f.Kind() == reflect.Ptr && f.IsNil() {
			str = nil
		} else {
			str, err = fs.String(f.Interface())
			if err != nil {
				return err
			}
		}

		val := "NULL"
		if str != nil {
			if !(isZero(*str) && fs.Schema.IsNullable) {
				val = fmt.Sprintf("%s", *str)
				if fs.Schema.IsEscape {
					val = fmt.Sprintf("%q", *str)
				}
			}
		}

		vals = append(vals, fmt.Sprintf("`%s` = %s", fs.Name, val))
	}

	if primaryKey.Incomplete() {
		return errors.New("goloquent: primary key not found")
	}

	cond := fmt.Sprintf(
		"`%s` = %q AND `%s` = %q",
		FieldNameKey, stringPrimaryKey(primaryKey),
		FieldNameParent, primaryKey.Parent.String())

	sql := fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE %s;",
		table, strings.Join(vals, ","), cond)

	if _, err := x.Exec(sql); err != nil {
		return err
	}

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

			strVal := "NULL"
			if !f.IsNil() {
				t := reflect.TypeOf(f.Interface())
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}

				parseFunc, isValid := interfaceToStringList[t]
				if !isValid {
					return ErrUnsupportDataType
				}

				str, _ := parseFunc(f.Interface())
				if str != nil {
					strVal = fmt.Sprintf("%s", *str)
					if !isNumber(t) {
						strVal = fmt.Sprintf("%q", *str)
					}
				}
			}

			vals = append(vals, fmt.Sprintf("`%s` = %s", key, strVal))
		}
	} else {
		// TODO: struct
		return fmt.Errorf("goloquent: unsupported data type %v", v)
	}

	if len(vals) <= 0 {
		return errors.New("goloquent: no update field provided")
	}

	sql := fmt.Sprintf("UPDATE `%s` SET %s", table, strings.Join(vals, ","))
	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if stmt.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", stmt.Limit)
	}
	sql += ";"

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// Delete :
func (x *SQLAdapter) Delete(query *Query, key *datastore.Key) error {
	table := key.Kind
	if query.table.name != "" {
		table = query.table.name
	}
	where := fmt.Sprintf(
		"WHERE `%s` = %q AND `%s` = %q",
		FieldNameKey, stringPrimaryKey(key), FieldNameParent, key.Parent.String())
	sql := fmt.Sprintf("DELETE FROM `%s` %s;", table, where)

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

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// SoftDelete :
func (x *SQLAdapter) SoftDelete(query *Query, key *datastore.Key) error {
	value := fmt.Sprintf("`%s` = %q", FieldNameSoftDelete, time.Now().UTC().Format(MySQLDateTimeFormat))
	where := fmt.Sprintf("WHERE `%s` = %q AND `%s` = %q",
		FieldNameKey, stringPrimaryKey(key), FieldNameParent, key.Parent.String())
	sql := fmt.Sprintf("UPDATE `%s` SET %s %s", query.table.name, value, where)

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
			dbName: x.dbName,
		},
	}
	return func(c *Connection, txn *sql.Tx) error {
		defer txn.Rollback()
		if err := callback(c); err != nil {
			return err
		}
		return txn.Commit()
	}(c, txn)
}
