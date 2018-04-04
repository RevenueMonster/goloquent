package goloquent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
)

var interfaceToStringList = map[reflect.Type]func(interface{}) (*string, error){
	typeOfString:       strToString,
	typeOfBool:         boolToString,
	typeOfInt:          intToString,
	typeOfInt8:         intToString,
	typeOfInt16:        intToString,
	typeOfInt32:        intToString,
	typeOfInt64:        int64ToString,
	typeOfFloat32:      float32ToString,
	typeOfFloat64:      float64ToString,
	typeOfByte:         byteToString,
	typeOfDataStoreKey: keyToString,
	typeOfTime:         timeToString,
	typeOfGeopoint:     geoPointToString,
}

// StructToMap : expose to public
func StructToMap(i interface{}) map[string]interface{} {
	return structToMap(i)
}

func geoPointToString(it interface{}) (*string, error) {
	b, err := json.Marshal(it)
	if err != nil {
		return nil, err
	}
	str := string(b)
	if str == "" {
		return nil, nil
	}
	return &str, nil
}

func isZero(i interface{}) bool {
	return reflect.DeepEqual(i, reflect.Zero(reflect.TypeOf(i)).Interface())
}

func sliceToInterface(val interface{}) ([]interface{}, error) {
	slice := make([]interface{}, 0)
	v := reflect.Indirect(reflect.ValueOf(val))
	t := v.Type().Elem()
	toStrFunc, isOther := interfaceToStringList[t]

	for i := 0; i < v.Len(); i++ {
		f := v.Index(i)
		it := f.Interface()
		if !isOther { // slice or struct TODO: support slice as well
			slice = append(slice, structToMap(it))
			continue
		}

		strValue, err := toStrFunc(it)
		if err != nil {
			return nil, err
		}
		slice = append(slice, strValue)
	}

	return slice, nil
}

func structToMap(val interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(val))
	m := make(map[string]interface{}, 0)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		isExported := (f.PkgPath == "")
		// Skip if not anonymous private property
		if !isExported && !f.Anonymous {
			continue
		}

		tag := newTag(f)
		// Skip if contain skip tag
		if tag.IsSkip() {
			continue
		}

		tagName := tag.Name

		iv := v.Field(i)
		it := iv.Interface()

		toStrFunc, isExist := interfaceToStringList[f.Type]
		if isExist {
			m[tagName], _ = toStrFunc(it)
			continue
		}

		t := f.Type
		switch t.Kind() {
		case reflect.Slice, reflect.Array:
			if t.Elem().Kind() == reflect.Struct {
				m[tagName], _ = sliceToInterface(it)
				continue
			}

			slice := make([]interface{}, iv.Len(), iv.Len())
			for j := range slice {
				slice[j] = iv.Index(j).Interface()
			}
			m[tagName] = slice
			continue

		case reflect.Struct:
			m[tagName] = structToMap(it)
			continue

		}

		m[tagName] = it
	}

	return m
}

func strToString(val interface{}) (*string, error) {
	str := val.(string)
	return &str, nil
}

func boolToString(val interface{}) (*string, error) {
	str := fmt.Sprintf("%t", val)
	return &str, nil
}

func intToString(val interface{}) (*string, error) {
	str := fmt.Sprintf("%d", val)
	return &str, nil
}

func int64ToString(val interface{}) (*string, error) {
	n := val.(int64)
	str := strconv.FormatInt(n, 10)
	return &str, nil
}

func float32ToString(val interface{}) (*string, error) {
	str := fmt.Sprintf("%.2f", val)
	return &str, nil
}

func float64ToString(val interface{}) (*string, error) {
	str := strconv.FormatFloat(val.(float64), 'f', 2, 64)
	return &str, nil
}

func byteToString(val interface{}) (*string, error) {
	str := base64.StdEncoding.EncodeToString(val.([]byte))
	return &str, nil
}

func keyToString(it interface{}) (*string, error) {
	if isZero(it) {
		return nil, nil
	}
	r := reflect.Indirect(reflect.ValueOf(it))
	key := r.Interface().(datastore.Key)
	str := key.String()
	if str == "" {
		return nil, nil
	}
	return &str, nil
}

func geoPointToMap(it interface{}) (map[string]float64, error) {
	m := make(map[string]float64, 0)
	v := reflect.ValueOf(it)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	gp := v.Interface().(datastore.GeoPoint)
	m["Lat"] = gp.Lat
	m["Lng"] = gp.Lng
	return m, nil
}

func timeToString(it interface{}) (*string, error) {
	str := it.(time.Time).Format(MySQLDateTimeFormat)
	if str == "" {
		return nil, nil
	}
	return &str, nil
}

func sliceToString(it interface{}) (*string, error) {
	t := reflect.TypeOf(it)
	if t == typeOfByte {
		str, err := byteToString(it)
		if err != nil {
			return nil, err
		}
		return str, nil
	}

	t = t.Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	i := make([]interface{}, 0)
	var (
		str *string
		err error
	)

	v := reflect.ValueOf(it)
	switch t {
	case typeOfTime:
		for j := 0; j < v.Len(); j++ {
			var dt *string
			f := v.Index(j)
			dt, err = timeToString(f.Interface().(time.Time))
			if err != nil {
				return nil, err
			}
			if dt != nil {
				i = append(i, *dt)
			}
		}

	case typeOfGeopoint:
		for j := 0; j < v.Len(); j++ {
			var gp map[string]float64
			f := v.Index(j)
			gp, err = geoPointToMap(f.Interface())
			if err != nil {
				return nil, err
			}
			i = append(i, gp)
		}

	case typeOfDataStoreKey:
		for j := 0; j < v.Len(); j++ {
			f := v.Index(j)
			key := new(datastore.Key)
			e := reflect.ValueOf(key).Elem()
			e.Set(reflect.Indirect(f))
			i = append(i, stringKey(key))
		}

	default:
		i, err = sliceToInterface(it)
		if err != nil {
			return str, err
		}

	}

	var b []byte
	b, err = json.Marshal(i)
	bStr := string(b)
	str = &bStr
	return str, nil
}

func structToString(it interface{}) (*string, error) {
	t := reflect.TypeOf(it)
	v := reflect.ValueOf(it)
	if it == nil || (t.Kind() == reflect.Ptr && v.IsNil()) {
		return nil, nil
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var (
		str *string
		err error
	)

	switch t {
	case typeOfSoftDelete:
		if it == nil {
			str = nil
			return str, nil
		}
		// str, err = timeToString(it)

	case typeOfTime:
		str, err = timeToString(it)

	case typeOfGeopoint:
		str, err = geoPointToString(it)

	case typeOfDataStoreKey:
		str, err = keyToString(it)

	default:
		m := structToMap(it)
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		bString := string(b)
		str = &bString

	}

	if err != nil {
		return nil, err
	}

	return str, nil
}
