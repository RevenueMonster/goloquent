package goloquent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/datastore"
)

var entityCacheList = map[string]*Entity{}

// Entity :
type Entity struct {
	columns     map[string]*Field
	fields      []*Field
	PrimaryKey  *Field
	Type        reflect.Type
	LoadKeyFunc func(interface{}, *datastore.Key) error
	LoadFunc    func(interface{}) error
	SaveFunc    func(interface{}) ([]datastore.Property, error)
}

// GetColumns :
func (e *Entity) GetColumns() map[string]*Field {
	return e.columns
}

// GetFields :
func (e *Entity) GetFields() []*Field {
	return e.fields
}

func getEntity(t reflect.Type) (*Entity, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	uniqueName := fmt.Sprintf("%s/%s", strings.TrimSpace(strings.Trim(t.PkgPath(), "/")), t.Name())
	if cache, isExist := entityCacheList[uniqueName]; isExist {
		return cache, nil
	}

	f, err := ListFields(t)
	if err != nil {
		return nil, err
	}
	loadKeyFunc := func(interface{}, *datastore.Key) error {
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
			if err := f.Load(props); err != nil {
				return err
			}
			return nil
		}
		saveFunc = func(i interface{}) ([]datastore.Property, error) {
			f, isOK := i.(datastore.KeyLoader)
			if isOK {
				return f.Save()
			}
			return nil, errors.New("invalid")
		}
	} else if r.Implements(propertyLoadInterface) {
		fmt.Println("is implement key loader")
		loadFunc = func(i interface{}) error {
			props := make([]datastore.Property, 0)
			f := i.(datastore.PropertyLoadSaver)
			if err := f.Load(props); err != nil {
				return err
			}
			return nil
		}
		saveFunc = func(i interface{}) ([]datastore.Property, error) {
			f, isOK := i.(datastore.PropertyLoadSaver)
			if isOK {
				return f.Save()
			}
			return nil, errors.New("invalid")
		}
	}

	cols := make(map[string]*Field, 0)
	for _, each := range f {
		cols[each.Name] = each
	}

	e := &Entity{
		Type:        t,
		LoadKeyFunc: loadKeyFunc,
		LoadFunc:    loadFunc,
		SaveFunc:    saveFunc,
		columns:     cols,
		fields:      f,
	}

	// Cache entity to avoid repeating reflect on the same struct
	entityCacheList[uniqueName] = e

	return e, nil
}
