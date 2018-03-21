package goloquent

import (
	"fmt"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
)

// StringPrimaryKey :
func StringPrimaryKey(key *datastore.Key) string {
	return stringPrimaryKey(key)
}

func stringPrimaryKey(key *datastore.Key) string {
	if key.Name != "" {
		return key.Name
	}

	return strconv.FormatInt(key.ID, 10)
}

func stringKey(key *datastore.Key) string {
	paths := make([]string, 0)
	for key != nil {
		paths = append(paths, strings.Trim(fmt.Sprintf("%v", key), "/"))
		key = key.Parent
	}
	return strings.Join(paths, "/")
}

// ParseKey :
func ParseKey(key string) (*datastore.Key, error) {
	strKey := strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidPrimaryKey
	}

	strKey = strings.Trim(strKey, "/")
	parents := strings.Split(strKey, "/")
	if len(parents) <= 0 {
		return nil, ErrInvalidPrimaryKey
	}

	parentKey := new(datastore.Key)
	for _, each := range parents {
		paths := strings.Split(each, ",")
		if len(paths) != 2 {
			return nil, ErrInvalidPrimaryKey
		}
		kind := paths[0]
		strID := paths[1]
		key := datastore.IncompleteKey(kind, nil)
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

	return parentKey, nil
}

func parsePrimaryKey(table string, id string, parent string) (*datastore.Key, error) {
	keyHandler := func(pk *datastore.Key) (*datastore.Key, error) {
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			key := datastore.NameKey(table, id, pk)
			return key, nil
		}
		key := datastore.IncompleteKey(table, pk)
		key.ID = intID
		return key, nil
	}

	parent = strings.TrimSpace(parent)
	if parent == "" {
		return keyHandler(nil)
	}

	parent = strings.Trim(parent, "/")
	parents := strings.Split(parent, "/")
	if len(parents) <= 0 {
		return nil, ErrParsePrimaryKey
	}

	parentKey := new(datastore.Key)
	for _, each := range parents {
		paths := strings.Split(each, ",")
		if len(paths) != 2 {
			return nil, ErrParsePrimaryKey
		}
		kind := paths[0]
		strID := paths[1]
		key := datastore.IncompleteKey(kind, nil)
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

	return keyHandler(parentKey)
}
