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
	sql := fmt.Sprintf(
		"SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = %q AND TABLE_NAME = %q;",
		x.dbName, table)

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(sql)

	if err != nil {
		return err
	}

	if len(results) > 0 {
		columnList := make(map[string]bool, 0)
		for _, item := range results {
			columnList[strings.ToLower(string(item["COLUMN_NAME"]))] = true
		}

		newCols := make([]*Field, 0)
		if entity.SoftDelete != nil {
			entity.SoftDelete.Name = FieldNameSoftDelete
			cols = append(cols, entity.SoftDelete)
		}
		for i, fs := range cols {
			_, isExist := columnList[strings.ToLower(fs.Name)]
			if !isExist {
				newCols = append(newCols, cols[i])
			}
		}

		script := x.toColumnSQL(newCols)
		if len(script) <= 0 {
			return nil
		}

		for i, item := range script {
			script[i] = fmt.Sprintf("ADD %s", item)
		}

		sql = fmt.Sprintf("ALTER TABLE `%s` %s;", table, strings.Join(script, ","))

		if _, err := x.Exec(sql); err != nil {
			return err
		}

		return nil
	}

	script := make([]string, 0)

	// Set primary key field
	script = append(script, fmt.Sprintf(
		"`%s` varchar(%d) CHARACTER SET `%s` COLLATE `%s` NOT NULL",
		FieldNameKey, IDLength, latin2CharSet.Encoding, latin2CharSet.Collation))
	script = append(script, fmt.Sprintf(
		"`%s` varchar(%d) CHARACTER SET `%s` COLLATE `%s` NOT NULL",
		FieldNameParent, KeyLength, latin2CharSet.Encoding, latin2CharSet.Collation))

	fieldScript := x.toColumnSQL(cols)
	script = append(script, fieldScript...)

	if entity.SoftDelete != nil {
		s := entity.SoftDelete.Schema
		script = append(script, fmt.Sprintf("`%s` %s", FieldNameSoftDelete, s.DataType))
	}

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
