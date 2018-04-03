package goloquent

// Limit :
type Limit struct {
	query *Query
}

func newLimit(q *Query) *Limit {
	return &Limit{q}
}

// Get :
func (l *Limit) Get(modelStruct interface{}) error {
	return newBuilder(l.query).Get(modelStruct)
}

// Update :
func (l *Limit) Update(values map[string]interface{}) error {
	return newBuilder(l.query).UpdateMulti(values)
}
