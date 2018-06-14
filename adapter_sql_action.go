package goloquent

import (
	"bytes"
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
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Zero(v.Type())
			}
			v = v.Elem()
		}
		v = v.Field(p)
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

	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("INSERT INTO %s ", quote(table)))
	buf.WriteString("(")
	buf.WriteString(quote(FieldNamePrimaryKey) + ",")
	buf.WriteString(quote(FieldNameKey) + ",")
	buf.WriteString(quote(FieldNameParent) + ",")
	kk, pp := stringPrimaryKey(primaryKey), primaryKey.Parent.String()
	args = append(args, pp+"/"+kk, kk, pp)
	for _, f := range cols {
		buf.WriteString(quote(f.Name) + ",")
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(") ")
	buf.WriteString(fmt.Sprintf("VALUES (%s);",
		strings.Trim(strings.Repeat("?,", len(cols)+3), ",")))

	// Run through every property in struct and convert to string
	v := reflect.ValueOf(modelStruct)
	if entity.PrimaryKey != nil {
		fv := v.Elem().FieldByIndex(entity.PrimaryKey.Index)
		fv.Set(reflect.ValueOf(primaryKey))
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

		var val interface{}
		if str != nil {
			if isZero(*str) && fs.Schema.IsNullable {
				val = nil
			} else {
				val = str
			}
		}

		args = append(args, val)
	}

	if _, err := x.Exec(buf.String(), args...); err != nil {
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
		if kv.Kind() == reflect.Slice {
			parent := kv.Index(i).Interface().(*datastore.Key)
			k = x.newPrimaryKey(table, parent)
		} else {
			switch kk := parentKey.(type) {
			case *datastore.Key:
				k = x.newPrimaryKey(table, kk)
			default:
				k = x.newPrimaryKey(table, nil)
			}
		}
		pKeys[i] = k
	}

	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("INSERT INTO %s ", quote(table)))
	buf.WriteString("(")
	buf.WriteString(quote(FieldNamePrimaryKey) + ",")
	buf.WriteString(quote(FieldNameKey) + ",")
	buf.WriteString(quote(FieldNameParent) + ",")
	for _, f := range cols {
		buf.WriteString(quote(f.Name) + ",")
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(") ")
	buf.WriteString("VALUES ")

	for i := 0; i < v.Len(); i++ {
		fv := v.Index(i)
		if !fv.IsValid() {
			return errors.New("goloquent (create multiple): invalid model data")
		}

		// Generate primary key before insert to database
		primaryKey := pKeys[i]
		buf.WriteString(fmt.Sprintf("(%s),", strings.Trim(strings.Repeat("?,", len(cols)+3), ",")))
		kk, pp := stringPrimaryKey(primaryKey), primaryKey.Parent.String()
		args = append(args, pp+"/"+kk, kk, pp)

		if entity.PrimaryKey != nil {
			vv := fv
			if fv.Kind() == reflect.Ptr {
				vv = vv.Elem()
			}
			ff := vv.FieldByIndex(entity.PrimaryKey.Index)
			ff.Set(reflect.ValueOf(primaryKey))
		}
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

			var val interface{}
			if str != nil {
				if isZero(*str) && fs.Schema.IsNullable {
					val = nil
				} else {
					val = str
				}
			}
			args = append(args, val)
		}
	}

	buf.Truncate(buf.Len() - 1)
	buf.WriteString(";")

	if _, err := x.Exec(buf.String(), args...); err != nil {
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
	v := reflect.ValueOf(modelStruct)
	if entity.PrimaryKey != nil {
		f := v.Elem().FieldByIndex(entity.PrimaryKey.Index)
		f.Set(reflect.ValueOf(primaryKey))
	}

	j := 3
	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("INSERT INTO %s ", quote(table)))
	buf.WriteString("(")
	buf.WriteString(quote(FieldNamePrimaryKey) + ",")
	buf.WriteString(quote(FieldNameKey) + ",")
	buf.WriteString(quote(FieldNameParent) + ",")
	// args = append(args, stringPrimaryKey(primaryKey), primaryKey.Parent.String())
	kk, pp := stringPrimaryKey(primaryKey), primaryKey.Parent.String()
	args = append(args, pp+"/"+kk, kk, pp)
	if entity.SoftDelete != nil {
		j++
		buf.WriteString(quote(FieldNameSoftDelete) + ",")
		f := v.Elem().FieldByIndex(entity.SoftDelete.Index)
		sd := f.Interface().(SoftDelete)
		var softDelete interface{}
		if !isZero(sd.DeletedDateTime) {
			softDelete = sd.DeletedDateTime.Format(MySQLDateTimeFormat)
		}
		args = append(args, softDelete)
	}
	onConflict := make([]string, 0, len(cols))
	for _, f := range cols {
		name := quote(f.Name)
		buf.WriteString(name + ",")
		if _, isOk := excludedFields[name]; !isOk {
			onConflict = append(onConflict, fmt.Sprintf("%s=VALUES(%s)", name, name))
		}
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(") ")
	buf.WriteString(fmt.Sprintf("VALUES (%s) ",
		strings.Trim(strings.Repeat("?,", len(cols)+j), ",")))
	buf.WriteString(fmt.Sprintf("ON DUPLICATE KEY UPDATE %s;",
		strings.Join(onConflict, ",")))

	for _, fs := range cols {
		f := getField(v.Elem(), fs.Index)
		if !f.IsValid() {
			return fmt.Errorf("goloquent: missing field on index %v", fs.Index)
		}

		str, err := fs.String(f.Interface())
		if err != nil {
			return err
		}

		var val interface{}
		if str != nil {
			if isZero(*str) && fs.Schema.IsNullable {
				val = nil
			} else {
				val = str
			}
		}
		args = append(args, val)
	}

	if _, err := x.Exec(buf.String(), args...); err != nil {
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
			switch kk := parentKey.(type) {
			case *datastore.Key:
				k = x.newPrimaryKey(table, kk)
			default:
				k = x.newPrimaryKey(table, nil)
			}
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

	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("INSERT INTO %s ", quote(table)))
	buf.WriteString("(")
	buf.WriteString(quote(FieldNamePrimaryKey) + ",")
	buf.WriteString(quote(FieldNameKey) + ",")
	buf.WriteString(quote(FieldNameParent) + ",")
	if entity.SoftDelete != nil {
		buf.WriteString(quote(FieldNameSoftDelete) + ",")
	}
	onConflict := make([]string, 0, len(cols))
	for _, f := range cols {
		name := quote(f.Name)
		buf.WriteString(name + ",")
		if _, isOk := excludedFields[name]; !isOk {
			onConflict = append(onConflict, fmt.Sprintf("%s=VALUES(%s)", name, name))
		}
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(") ")
	buf.WriteString("VALUES ")

	for i := 0; i < v.Len(); i++ {
		fv := v.Index(i)
		if !fv.IsValid() {
			return errors.New("goloquent: invalid model data")
		}

		// Generate primary key before insert to database
		primaryKey := pKeys[i]
		if entity.PrimaryKey != nil {
			ff := getField(fv, entity.PrimaryKey.Index)
			ff.Set(reflect.ValueOf(primaryKey))
		}

		j := 3
		// args = append(args, stringPrimaryKey(primaryKey), primaryKey.Parent.String())
		kk, pp := stringPrimaryKey(primaryKey), primaryKey.Parent.String()
		args = append(args, pp+"/"+kk, kk, pp)
		if entity.SoftDelete != nil {
			j++
			f := v.Elem().FieldByIndex(entity.SoftDelete.Index)
			sd := f.Interface().(SoftDelete)
			var softDelete interface{}
			if !isZero(sd.DeletedDateTime) {
				softDelete = sd.DeletedDateTime.Format(MySQLDateTimeFormat)
			}
			args = append(args, softDelete)
		}
		buf.WriteString(fmt.Sprintf("(%s),",
			strings.Trim(strings.Repeat("?,", len(cols)+j), ",")))

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

			var val interface{}
			if str != nil {
				if isZero(*str) && fs.Schema.IsNullable {
					val = nil
				} else {
					val = str
				}
			}
			args = append(args, val)
		}
	}

	buf.Truncate(buf.Len() - 1)
	buf.WriteString(fmt.Sprintf(" ON DUPLICATE KEY UPDATE %s;",
		strings.Join(onConflict, ",")))

	if _, err := x.Exec(buf.String(), args...); err != nil {
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
	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("UPDATE %s SET ", quote(table)))

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

		var val interface{}
		if str != nil {
			if isZero(*str) && fs.Schema.IsNullable {
				val = nil
			} else {
				val = str
			}
		}
		buf.WriteString(fmt.Sprintf("%s = ?,", quote(fs.Name)))
		args = append(args, val)
	}

	if primaryKey.Incomplete() {
		return errors.New("goloquent: primary key not found")
	}

	buf.Truncate(buf.Len() - 1)
	buf.WriteString(fmt.Sprintf(" WHERE %s = ?",
		quote(FieldNamePrimaryKey)))
	buf.WriteString(";")
	pp, kk := primaryKey.Parent.String(), stringPrimaryKey(primaryKey)
	args = append(args, pp+"/"+kk)

	if _, err := x.Exec(buf.String(), args...); err != nil {
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

	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("UPDATE %s SET ", quote(table)))
	v := reflect.Indirect(reflect.ValueOf(modelStruct))
	if v.Kind() == reflect.Map {
		if v.Len() <= 0 {
			return errors.New("goloquent: no update field provided")
		}

		for _, key := range v.MapKeys() {
			f := v.MapIndex(key)

			var vv interface{}
			if !f.IsNil() {
				t := reflect.TypeOf(f.Interface())
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}

				parseFunc, isValid := interfaceToStringList[t]
				if !isValid {
					return ErrUnsupportDataType
				}

				vv, _ = parseFunc(f.Interface())
			}

			buf.WriteString(fmt.Sprintf("%s = ?,", quote(key.String())))
			args = append(args, vv)
		}
	} else {
		return fmt.Errorf("goloquent: unsupported data type %v", v)
	}
	buf.Truncate(buf.Len() - 1)

	if len(stmt.Where) > 0 {
		buf.WriteString(fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND ")))
	}
	if stmt.Limit > 0 {
		buf.WriteString(fmt.Sprintf(" LIMIT %d", stmt.Limit))
	}
	buf.WriteString(";")
	if _, err := x.Exec(buf.String(), args...); err != nil {
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
	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("DELETE FROM %s ", quote(table)))
	buf.WriteString(fmt.Sprintf(
		"WHERE %s = ?",
		quote(FieldNamePrimaryKey)))
	buf.WriteString(";")
	args = append(args, key.Parent.String()+"/"+stringPrimaryKey(key))
	if _, err := x.Exec(buf.String(), args...); err != nil {
		return err
	}
	return nil
}

// DeleteMulti :
func (x *SQLAdapter) DeleteMulti(query *Query, keys []*datastore.Key) error {
	args := make([]interface{}, 0)
	for _, key := range keys {
		args = append(args, key.Parent.String()+"/"+stringPrimaryKey(key))
	}

	table := query.table.name
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("DELETE FROM %s ", quote(table)))
	buf.WriteString(fmt.Sprintf("WHERE %s IN (%s)",
		quote(FieldNamePrimaryKey),
		strings.Trim(strings.Repeat("?,", len(args)), ",")))

	if _, err := x.Exec(buf.String(), args...); err != nil {
		return err
	}

	return nil
}

// SoftDelete :
func (x *SQLAdapter) SoftDelete(query *Query, key *datastore.Key) error {
	buf, args := new(bytes.Buffer), make([]interface{}, 0)
	buf.WriteString(fmt.Sprintf("UPDATE %s SET %s = ? ",
		quote(query.table.name),
		quote(FieldNameSoftDelete)))
	buf.WriteString(fmt.Sprintf("WHERE %s = ? AND %s = ?;",
		quote(FieldNameKey),
		quote(FieldNameParent)))
	args = append(args,
		time.Now().UTC().Format(MySQLDateTimeFormat),
		stringPrimaryKey(key),
		key.Parent.String())

	if _, err := x.Exec(buf.String(), args...); err != nil {
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
