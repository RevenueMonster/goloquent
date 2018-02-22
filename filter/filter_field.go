package filter

// string = like, eq, ne, in, nin
// int, int8, int16, int18, int32, int64
// uint, uint8, uint16, uint32, uint64
// 	= eq, ne, gt, gte, lt, lte, in, nin
// float32, float64 = eq, ne, gt, gte, lt, lte, in, nin
// []byte = like, eq, ne, in, nin
// bool = eq, ne
// time = eq, ne, gt, gte, lt, lte, in, nin
// (#) Geopoint = eq, ne, in, nin (not support)
// (#) Key = eq, ne, gt, gte, lt, lte, in, nin, like (not support?!)

// Field :
type Field struct {
	Name     string
	JSONName string
	mapFunc  filterFunc
}

// Map :
func (f *Field) Map(it interface{}) ([]Filter, error) {
	return f.mapFunc(f, it)
}

func newField(tag *Tag, mapFunc filterFunc) *Field {
	return &Field{
		Name:     tag.Name,
		JSONName: tag.JSONName,
		mapFunc:  mapFunc,
	}
}
