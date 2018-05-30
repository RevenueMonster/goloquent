package goloquent

// Pagination :
type Pagination struct {
	Cursor  string      `json:"cursor,omitempty" form:"cursor"`
	Filter  interface{} `json:"filter,omitempty" form:"filter"`
	OrderBy []string    `json:"order,omitempty" form:"order"`
	Count   uint        `json:"-"`
	Limit   uint        `json:"-"`
	// Total   uint        `json:"-"`
}
