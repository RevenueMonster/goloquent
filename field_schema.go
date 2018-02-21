package goloquent

// CharSet :
type CharSet struct {
	Encoding  string
	Collation string
}

// FieldSchema :
type FieldSchema struct {
	DataType     string
	DefaultValue interface{}
	IsEscape     bool
	IsUnique     bool
	IsUnsigned   bool
	IsNullable   bool
	*CharSet
}
