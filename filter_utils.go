package goloquent

import (
	"reflect"
	"strings"
)

// FilterScan :
type FilterScan struct {
	JSONName []string
	StructScan
}

func listFilter(it interface{}) (map[string]*FilterField, error) {
	t := reflect.TypeOf(it)

	fieldList := make(map[string]*FilterField, 0)
	scanStructs := make([]*FilterScan, 0)
	scanStructs = append(scanStructs, &FilterScan{
		JSONName: make([]string, 0),
		StructScan: StructScan{
			Type:   t,
			Column: make([]string, 0),
			Index:  make([]int, 0),
		},
	})

	for len(scanStructs) > 0 {
		// Get the first element
		r := scanStructs[0]
		for i := 0; i < r.Type.NumField(); i++ {
			f := r.Type.Field(i)

			isExported := (f.PkgPath == "")

			// Skip if not anonymous private property
			if !isExported && !f.Anonymous {
				continue
			}

			tag := newFilterTag(f)
			if tag.JSONName == "-" {
				continue
			}

			name := make([]string, 0)
			name = append(name, r.Column...)
			name = append(name, tag.Name)

			jsonName := make([]string, 0)
			jsonName = append(jsonName, r.JSONName...)
			jsonName = append(jsonName, tag.JSONName)

			index := make([]int, 0)
			index = append(index, r.Index...)
			index = append(index, i)

			nameKey := strings.Join(jsonName, ".")

			t := f.Type
			mapFunc, err := isValidType(t)
			if err != nil {
				return nil, err
			}

			if mapFunc != nil {
				fieldList[nameKey] = newFilterField(tag, mapFunc)
				continue
			}

			// if it's embedded struct
			if f.Anonymous {
				// Skip if embedded struct is unexported
				if !isExported {
					continue
				}

				scanStructs = append(scanStructs, &FilterScan{
					JSONName: make([]string, 0),
					StructScan: StructScan{
						Type:   t,
						Column: make([]string, 0),
						Index:  index,
					},
				})
				continue
			}

			scanStructs = append(scanStructs, &FilterScan{
				JSONName: jsonName,
				StructScan: StructScan{
					Type:   t,
					Column: name,
					Index:  index,
				},
			})
		}
		// unshift scan struct
		scanStructs = scanStructs[1:]
	}

	return fieldList, nil
}
