package goloquent

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

// GeoPoint :
type GeoPoint struct {
	Lat json.Number
	Lng json.Number
}

func mapToStruct(t reflect.Type, m map[string]interface{}) (interface{}, error) {
	isPtr := false
	if t.Kind() == reflect.Ptr {
		if len(m) == 0 {
			return nil, nil
		}
		isPtr = true
		t = t.Elem()
	}

	vi := reflect.New(t)
	for k, val := range m {
		f := reflect.Indirect(vi).FieldByName(k)
		if !f.IsValid() {
			continue
		}

		switch f.Kind() {
		case reflect.Float32, reflect.Float64:
			var it float64
			switch vi := val.(type) {
			case int64:
				it = float64(vi)
			case string:
				it, _ = strconv.ParseFloat(vi, 64)
			case []byte:
				it, _ = strconv.ParseFloat(string(vi), 64)
			}
			f.SetFloat(it)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var it int64
			switch vi := val.(type) {
			case float64:
				it = int64(vi)
			case string:
				it, _ = strconv.ParseInt(vi, 10, 64)
			case []byte:
				it, _ = strconv.ParseInt(string(vi), 10, 64)
			}
			f.SetInt(it)

		default:
			f.Set(reflect.ValueOf(val))
		}
	}

	if !isPtr {
		vi = vi.Elem()
	}

	return vi.Interface(), nil
}

func stringToStr(f *Field, val string) (interface{}, error) {
	return val, nil
}

func stringToBool(f *Field, val string) (interface{}, error) {
	return strconv.ParseBool(val)
}

func stringToInt(f *Field, val string) (interface{}, error) {
	i, err := strconv.Atoi(val)
	return i, err
}

func stringToInt8(f *Field, val string) (interface{}, error) {
	i, err := strconv.Atoi(val)
	return int8(i), err
}

func stringToInt16(f *Field, val string) (interface{}, error) {
	i, err := strconv.Atoi(val)
	return int16(i), err
}

func stringToInt32(f *Field, val string) (interface{}, error) {
	i, err := strconv.Atoi(val)
	return int32(i), err
}

func stringToInt64(f *Field, val string) (interface{}, error) {
	return strconv.ParseInt(val, 10, 64)
}

func stringToFloat32(f *Field, val string) (interface{}, error) {
	number, err := strconv.ParseFloat(val, 32)
	return float32(number), err
}

func stringToFloat64(f *Field, val string) (interface{}, error) {
	return strconv.ParseFloat(val, 64)
}

func stringToByte(f *Field, val string) (interface{}, error) {
	return base64.StdEncoding.DecodeString(val)
}

func stringToKey(val string) (*datastore.Key, error) {
	path := strings.TrimSpace(strings.Trim(val, "/"))
	paths := strings.Split(path, "/")
	if len(paths) <= 0 {
		return nil, ErrInvalidPrimaryKey
	}

	parentKey := new(datastore.Key)
	key := new(datastore.Key)
	for _, p := range paths {
		part := strings.Split(p, ",")
		if len(part) != 2 {
			return nil, ErrInvalidPrimaryKey
		}

		kind := part[0]
		strID := part[1]
		key = datastore.IncompleteKey(kind, nil)
		i, err := strconv.ParseInt(strID, 10, 64)
		if err != nil {
			key.Name = strID
		} else {
			key.ID = i
		}
		if !parentKey.Incomplete() {
			key.Parent = parentKey
		}
		parentKey = key
	}

	return key, nil
}

func stringToGeoPoint(val string) (*datastore.GeoPoint, error) {
	gp := new(datastore.GeoPoint)
	if err := json.Unmarshal([]byte(val), gp); err != nil {
		return nil, err
	}
	return gp, nil
}

func stringToTime(val string) (interface{}, error) {
	return time.Parse(MySQLDateTimeFormat, val)
}

func stringToStruct(f *Field, val string) (interface{}, error) {
	if val == "" {
		return nil, nil
	}
	t := f.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var (
		it  interface{}
		err error
	)

	switch t {
	case typeOfTime:
		it, err = stringToTime(val)

	case typeOfGeopoint:
		it, err = stringToGeoPoint(val)
		if f.Type.Kind() != reflect.Ptr {
			it = *it.(*datastore.GeoPoint)
		}

	case typeOfDataStoreKey:
		it, err = stringToKey(val)
		if f.Type.Kind() != reflect.Ptr {
			it = *it.(*datastore.Key)
		}

	default:
		if val == "" {
			return reflect.New(f.Type).Elem().Interface(), nil
		}
		m := make(map[string]interface{}, 0)
		if err := json.Unmarshal([]byte(val), &m); err != nil {
			return nil, err
		}
		it, err = mapToStruct(f.Type, m)

	}

	if err != nil {
		return nil, err
	}

	return it, nil
}

func unmarshalToSliceString(i string) ([]string, error) {
	s := make([]string, 0)
	if err := json.Unmarshal([]byte(i), &s); err != nil {
		return nil, err
	}
	return s, nil
}

func stringToSlice(f *Field, i string) (interface{}, error) {
	if f.Type == typeOfByte {
		return stringToByte(f, i)
	}

	t := f.Type.Elem()
	isPtr := (t.Kind() == reflect.Ptr)
	if isPtr {
		t = t.Elem()
	}

	var it interface{}
	switch t {
	case typeOfTime:
		s, err := unmarshalToSliceString(i)
		if err != nil {
			return nil, err
		}
		times := make([]time.Time, 0)
		for _, item := range s {
			dt, err := time.Parse(MySQLDateTimeFormat, item)
			if err != nil {
				return nil, err
			}
			times = append(times, dt)
		}
		it = times

	case typeOfGeopoint:
		cols := make([]GeoPoint, 0)
		err := json.Unmarshal([]byte(i), &cols)
		if err != nil {
			return nil, err
		}
		slice := reflect.MakeSlice(f.Type, len(cols), len(cols))
		for j, item := range cols {
			lat, _ := item.Lat.Float64()
			lng, _ := item.Lng.Float64()
			geo := &datastore.GeoPoint{
				Lat: lat,
				Lng: lng,
			}
			vg := reflect.ValueOf(geo)
			if !isPtr {
				vg = vg.Elem()
			}
			slice.Index(j).Set(vg)
		}
		it = slice.Interface()

	case typeOfDataStoreKey:
		s, err := unmarshalToSliceString(i)
		if err != nil {
			return nil, err
		}

		r := reflect.New(f.Type)
		keys := r.Elem()
		for _, item := range s {
			k, err := stringToKey(item)
			if err != nil {
				return nil, err
			}

			v := reflect.ValueOf(k)
			if !isPtr {
				v = reflect.Indirect(v)
			}
			keys = reflect.Append(keys, v)
		}
		it = keys.Interface()

	default:
		if i == "" {
			return reflect.MakeSlice(f.Type, 0, 0).Interface(), nil
		}
		if f.Type.Kind() == reflect.Struct {
			// var s []json.RawMessage
			// if err := json.Unmarshal([]byte(i), s); err != nil {
			// 	return nil, err
			// }
			// for _, vv := range s {
			// 	stringToStruct(f.)
			// }
		}
		s := reflect.New(f.Type)
		if err := json.Unmarshal([]byte(i), s.Interface()); err != nil {
			return nil, err
		}
		it = s.Elem().Interface()
	}

	return it, nil
}
