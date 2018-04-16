package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"cloud.google.com/go/datastore"
)

var entityList sync.Map

// Entity :
type Entity struct {
	name        string
	fields      []*Field
	PrimaryKey  *Field
	Type        reflect.Type
	SoftDelete  *Field
	loadKeyFunc func(interface{}, *datastore.Key) error
	LoadFunc    func(interface{}) error
	SaveFunc    func(interface{}) ([]datastore.Property, error)
}

// GetFields :
func (e *Entity) GetFields() []*Field {
	return e.fields
}

// LoadKey :
func (e *Entity) LoadKey(i interface{}, key *datastore.Key) error {
	// Skip to load key if model doesn't have key field
	if e.PrimaryKey == nil {
		return nil
	}

	return e.loadKeyFunc(i, key)
}

func getEntity(t reflect.Type) (*Entity, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	uniqueName := fmt.Sprintf("%s/%s", strings.TrimSpace(strings.Trim(t.PkgPath(), "/")), t.Name())
	if cache, isExist := entityList.Load(uniqueName); isExist {
		return cache.(*Entity), nil
	}

	k, s, f, err := ListFields(t)
	if err != nil {
		return nil, err
	}

	loadKeyFunc := func(i interface{}, key *datastore.Key) error {
		v := reflect.ValueOf(i)
		f := v.Elem().FieldByIndex(k.Index)
		if !f.IsValid() {
			return nil
		}
		f.Set(reflect.ValueOf(key))
		return nil
	}

	loadFunc := func(interface{}) error {
		return nil
	}
	saveFunc := func(interface{}) ([]datastore.Property, error) {
		return []datastore.Property{}, nil
	}

	i := reflect.New(t).Interface()
	r := reflect.TypeOf(i)
	keyLoadInterface := reflect.TypeOf((*datastore.KeyLoader)(nil)).Elem()
	propertyLoadInterface := reflect.TypeOf((*datastore.PropertyLoadSaver)(nil)).Elem()
	if r.Implements(keyLoadInterface) {
		loadKeyFunc = func(i interface{}, key *datastore.Key) error {
			f := i.(datastore.KeyLoader)
			return f.LoadKey(key)
		}
		loadFunc = func(i interface{}) error {
			props := make([]datastore.Property, 0)
			f := i.(datastore.KeyLoader)
			return f.Load(props)
		}
		saveFunc = func(i interface{}) ([]datastore.Property, error) {
			f, isOK := i.(datastore.KeyLoader)
			if isOK {
				return f.Save()
			}
			return nil, errors.New("invalid")
		}
	} else if r.Implements(propertyLoadInterface) {
		loadFunc = func(i interface{}) error {
			props := make([]datastore.Property, 0)
			f := i.(datastore.PropertyLoadSaver)
			return f.Load(props)
		}
		saveFunc = func(i interface{}) ([]datastore.Property, error) {
			f, isOK := i.(datastore.PropertyLoadSaver)
			if isOK {
				return f.Save()
			}
			return nil, errors.New("invalid")
		}
	}

	e := &Entity{
		name:        name,
		fields:      f,
		Type:        t,
		PrimaryKey:  k,
		SoftDelete:  s,
		loadKeyFunc: loadKeyFunc,
		LoadFunc:    loadFunc,
		SaveFunc:    saveFunc,
	}

	// Cache entity to avoid repeating reflect on the same struct
	entityList.Store(uniqueName, e)

	return e, nil
}
