package goloquent

import (
	"fmt"
	"strconv"
	"strings"
)

// Count :
func (x *SQLAdapter) Count(query *Query) (uint, error) {
	stmt, err := x.CompileStatement(query)
	if err != nil {
		return 0, err
	}

	key := "COUNT(`$Key`)"
	sql := fmt.Sprintf("SELECT %s FROM (%s) AS `Master`", key, strings.Join(stmt.Table, " UNION ALL "))

	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if len(stmt.Order) > 0 {
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(stmt.Order, ","))
	}

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(sql)
	if err != nil {
		return 0, err
	}

	intCount := int(0)
	for _, r := range results {
		if bc, isExist := r[key]; isExist {
			intCount, _ = strconv.Atoi(string(bc))
			break
		}
	}

	return uint(intCount), nil
}

// Sum :
func (x *SQLAdapter) Sum(field string, query *Query) (int, error) {
	table := query.table.name
	stmt, err := x.CompileStatement(query)
	if err != nil {
		return 0, err
	}

	key := fmt.Sprintf("COALESCE(SUM(`%s`),0)", field)
	sql := fmt.Sprintf("SELECT %s FROM `%s`", key, table)

	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}

	results := make([]map[string][]byte, 0)
	results, err = x.ExecQuery(sql)
	if err != nil {
		return 0, err
	}

	intSum := int(0)
	for _, r := range results {
		if bc, isExist := r[key]; isExist {
			intSum, _ = strconv.Atoi(string(bc))
			break
		}
	}

	return intSum, nil
}
