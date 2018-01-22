package goloquent

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
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
	// TODO: add dbname as prefix
	// sql := fmt.Sprintf("SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = %q;", table)
	// results := make([]map[string][]byte, 0)
	// results, err = x.ExecQuery(sql)

	// if len(results) > 0 {
	// 	script := make([]string, 0)

	// 	existed := make(map[string]bool, 0)
	// 	for _, data := range results {
	// 		k := string(data["COLUMN_NAME"])
	// 		k = strings.ToLower(k)
	// 		existed[k] = true
	// 	}

	// 	fieldScript := x.toColumnSql(cols)
	// 	script = append(script, fieldScript...)

	// 	if len(script) <= 0 {
	// 		return nil
	// 	}

	// 	for i, each := range script {
	// 		script[i] = fmt.Sprintf("ADD %s", each)
	// 	}

	// 	fmt.Println(existed)

	// 	sql = fmt.Sprintf("ALTER TABLE %s %s;", table, strings.Join(script, ","))

	// 	fmt.Println("************* START MIGRATION ALTER QUERY ************")
	// 	fmt.Println(color.GreenString(sql))
	// 	fmt.Println("************* ENDED MIGRATION ALTER QUERY ************")

	// 	if _, err := x.Exec(sql); err != nil {
	// 		return err
	// 	}

	// 	return nil
	// }

	script := make([]string, 0)

	// Set primary key field
	script = append(script, fmt.Sprintf(
		"`%s` varchar(%d) CHARACTER SET `%s` COLLATE `%s` NOT NULL",
		FieldNameKey, IDLength, latin2CharSet.Encoding, latin2CharSet.Collation))
	script = append(script, fmt.Sprintf(
		"`%s` varchar(%d) CHARACTER SET `%s` COLLATE `%s` NOT NULL",
		FieldNameParent, KeyLength, latin2CharSet.Encoding, latin2CharSet.Collation))

	fieldScript := x.toColumnSql(cols)
	script = append(script, fieldScript...)

	// Index primary key field
	script = append(script, fmt.Sprintf(
		"CONSTRAINT `%s` UNIQUE (`%s`, `%s`)",
		FieldNamePrimaryKey, FieldNameKey, FieldNameParent))

	sql := fmt.Sprintf(
		"CREATE TABLE `%s` (%s) CHARACTER SET `%s` COLLATE `%s`;",
		table, strings.Join(script, ","), utf8CharSet.Encoding, utf8CharSet.Collation)

	fmt.Println("************* START MIGRATION QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED MIGRATION QUERY ************")

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// Drop :
func (x *SQLAdapter) Drop(query *Query) error {
	sql := fmt.Sprintf("DROP TABLE `%s`", query.table.name)

	fmt.Println("************* START DROP TABLE QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED DROP TABLE QUERY ************")

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// DropIfExists :
func (x *SQLAdapter) DropIfExists(query *Query) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", query.table.name)

	fmt.Println("************* START DROP TABLE IF EXISTS QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED DROP TABLE IF EXISTS QUERY ************")

	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}

// UniqueIndex :
func (x *SQLAdapter) UniqueIndex(query *Query, fields []string) error {
	table := query.table.name
	sql := fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON `%s` (%s);", strings.Join(fields, "_"), table, strings.Join(fields, ","))

	fmt.Println("************* START DROP TABLE IF EXISTS QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED DROP TABLE IF EXISTS QUERY ************")
	if _, err := x.Exec(sql); err != nil {
		return err
	}

	return nil
}
