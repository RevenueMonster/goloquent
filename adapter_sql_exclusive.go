package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Migrate : create table base on the struct schema
func (x *SQLAdapter) Migrate(query *Query, modelStruct interface{}) error {
	t := reflect.TypeOf(modelStruct)

	table := query.table.name
	entity, err := getEntity(t)
	if err != nil {
		return err
	}

	cols := entity.GetFields()
	columns := make([]*Field, 0)
	columns = append(columns, &Field{FieldNameKey, FieldNameKey, true, false, nil, nil,
		&FieldSchema{fmt.Sprintf("varchar(%d)", IDLength), nil, true, false, false, false, latin2CharSet}, nil})
	columns = append(columns, &Field{FieldNameParent, FieldNameParent, true, false, nil, nil,
		&FieldSchema{fmt.Sprintf("varchar(%d)", KeyLength), nil, true, false, false, false, latin2CharSet}, nil})
	columns = append(columns, cols...)

	if entity.SoftDelete != nil {
		columns = append(columns, entity.SoftDelete)
	}

	sql := fmt.Sprintf(
		"SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = %q AND TABLE_NAME = %q;",
		x.dbName, table)

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(sql)

	if err != nil {
		return err
	}

	if len(results) > 0 {
		columnList := make(map[string]int, 0)
		for i, item := range results {
			columnList[string(item["COLUMN_NAME"])] = i
		}

		syncColumns := make([]*Field, 0)
		newColumns := make([]*Field, 0)
		delColumns := make([]string, 0)

		positionList := make(map[string]string, 0)
		for i, fs := range columns {
			_, isExist := columnList[fs.Name]
			name := ""
			if i > 0 {
				name = (columns[i-1]).Name
			}
			positionList[fs.Name] = name
			if isExist {
				syncColumns = append(syncColumns, fs)
				delete(columnList, fs.Name)
				continue
			}
			newColumns = append(newColumns, columns[i])
		}

		// Get those deprecated columns
		for k := range columnList {
			delColumns = append(delColumns, k)
		}

		sql = fmt.Sprintf("ALTER TABLE `%s`", table)
		stmt := make([]string, 0)
		if len(newColumns) > 0 {
			script := x.toSQLSchema(newColumns)
			for i, item := range newColumns {
				name := positionList[item.Name]
				suffix := "FIRST"
				if name != "" {
					suffix = fmt.Sprintf("AFTER `%s`", name)
				}
				script[i] = fmt.Sprintf("ADD %s %s", script[i], suffix)
			}
			stmt = append(stmt, strings.Join(script, ","))
		}

		if len(syncColumns) > 0 {
			script := x.toSQLSchema(syncColumns)
			for i, item := range syncColumns {
				name := positionList[item.Name]
				suffix := "FIRST"
				if name != "" {
					suffix = fmt.Sprintf("AFTER `%s`", name)
				}
				script[i] = fmt.Sprintf("MODIFY %s %s", script[i], suffix)
			}
			stmt = append(stmt, strings.Join(script, ","))
		}

		if len(delColumns) > 0 {
			for i, item := range delColumns {
				delColumns[i] = fmt.Sprintf("DROP `%s`", item)
			}
			stmt = append(stmt, strings.Join(delColumns, ","))
		}
		sql += " " + strings.Join(stmt, ",")
		sql += ";"

		if _, err := x.Exec(sql); err != nil {
			return err
		}

		return nil
	}

	script := make([]string, 0)

	fieldScript := x.toSQLSchema(columns)
	script = append(script, fieldScript...)

	// Index primary key field
	script = append(script, fmt.Sprintf(
		"CONSTRAINT `%s` UNIQUE (`%s`, `%s`)",
		FieldNamePrimaryKey, FieldNameParent, FieldNameKey))

	sql = fmt.Sprintf(
		"CREATE TABLE `%s` (%s) CHARACTER SET `%s` COLLATE `%s`;",
		table, strings.Join(script, ","), utf8CharSet.Encoding, utf8CharSet.Collation)

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// Drop :
func (x *SQLAdapter) Drop(query *Query) error {
	sql := fmt.Sprintf("DROP TABLE `%s`", query.table.name)

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// DropIfExists :
func (x *SQLAdapter) DropIfExists(query *Query) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", query.table.name)

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// DropUniqueIndex :
func (x *SQLAdapter) DropUniqueIndex(query *Query, fields ...string) error {
	table := query.table.name

	sql := fmt.Sprintf(
		"SELECT * FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = %q AND TABLE_NAME = %q AND INDEX_NAME = %q;",
		x.dbName, table, strings.Join(fields, "_"))
	results, err := x.ExecQuery(sql)
	if err != nil {
		return err
	}

	if len(results) > 0 {
		sql = fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP INDEX `%s`;", x.dbName, table, strings.Join(fields, "_"))
		if _, err := x.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

// UniqueIndex :
func (x *SQLAdapter) UniqueIndex(query *Query, fields ...string) error {
	table := query.table.name

	sql := fmt.Sprintf(
		"SELECT * FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = %q AND TABLE_NAME = %q AND INDEX_NAME = %q;",
		x.dbName, table, strings.Join(fields, "_"))
	results, err := x.ExecQuery(sql)
	if err != nil {
		return err
	}

	if len(results) > 0 {
		sql = fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP INDEX `%s`;", x.dbName, table, strings.Join(fields, "_"))
		if _, err := x.Exec(sql); err != nil {
			return err
		}
	}

	sql = fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON `%s`.`%s` (%s);", strings.Join(fields, "_"), x.dbName, table, strings.Join(fields, ","))
	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// deleteWithQuery :
func (x *SQLAdapter) deleteWithQuery(query *Query) error {
	table := query.table.name
	stmt, err := x.CompileStatement(query)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("DELETE FROM `%s`.`%s` WHERE ", x.dbName, table)

	if len(stmt.Where) <= 0 {
		return errors.New("goloquent: delete statement without where statement is not allow")
	}

	sql += fmt.Sprintf("%s", strings.Join(stmt.Where, " AND "))

	_, err = x.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}
