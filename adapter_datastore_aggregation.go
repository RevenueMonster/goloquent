package goloquent

// Count :
func (ds *DataStoreAdapter) Count(query *Query) (uint, error) {
	q, err := ds.CompileQuery(query)
	if err != nil {
		return 0, err
	}

	c, err := ds.client.Count(ds.context, q)
	if err != nil {
		return 0, err
	}

	return uint(c), nil
}
