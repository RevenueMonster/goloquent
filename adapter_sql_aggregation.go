package goloquent

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// Count :
func (x *SQLAdapter) Count(query *Query) (int, error) {
	table := query.table.name
	stmt, err := x.CompileStatement(query)
	if err != nil {
		return 0, err
	}

	key := fmt.Sprintf("COUNT(`%s`)", FieldNameKey)
	sql := fmt.Sprintf("SELECT %s FROM `%s`", key, table)

	if len(stmt.Where) > 0 {
		sql += fmt.Sprintf(" WHERE %s", strings.Join(stmt.Where, " AND "))
	}
	if len(stmt.Order) > 0 {
		sql += fmt.Sprintf(" ORDER BY %s", strings.Join(stmt.Order, ","))
	}

	fmt.Println("************* START COUNT QUERY ************")
	fmt.Println(color.GreenString(sql))
	fmt.Println("************* ENDED COUNT QUERY ************")

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

	return intCount, nil
}
