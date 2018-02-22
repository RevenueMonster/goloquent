package filter


func list(input interface{}, b []byte) ([]Filter, error) {

	fieldList, err := getFilter(input)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{}, 0)
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	filters := make([]Filter, 0)
	for k, item := range m {
		field, isExist := fieldList[k]
		if !isExist {
			return nil, fmt.Errorf("filterable : invalid json key %q", k)
		}
		f, err := field.Map(item)
		if err != nil {
			return nil, err
		}
		filters = append(filters, f...)
	}

	return filters, nil
}