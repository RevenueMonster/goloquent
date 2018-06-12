package goloquent

import "strings"

// Dictionary :
type dictionary map[string]bool

func newDictionary(v []string) dictionary {
	l := make(map[string]bool)
	for _, vv := range v {
		vv = strings.TrimSpace(vv)
		if vv == "" {
			continue
		}
		l[vv] = true
	}
	return dictionary(l)
}

func (d dictionary) add(k string) {
	if !d.has(k) {
		d[k] = true
	}
}

// Has :
func (d dictionary) has(k string) bool {
	return d[k]
}

// Delete :
func (d dictionary) delete(k string) {
	delete(d, k)
}

// Keys :
func (d dictionary) keys() []string {
	arr := make([]string, 0, len(d))
	for k := range d {
		arr = append(arr, k)
	}
	return arr
}
