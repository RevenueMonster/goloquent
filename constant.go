package goloquent

import (
	"errors"
	"reflect"
	"time"

	"cloud.google.com/go/datastore"
)

// public constant variable
const (
	FieldNameKey        = "$Key"
	FieldNameParent     = "$Parent"
	FieldNamePrimaryKey = "$PrimaryKey"
	FieldNameSoftDelete = "DeletedAt"
	DefaultTotalRecord  = uint(100)
	IDLength            = 20
	KeyLength           = 767
	TextLength          = 255
	MaxKeyLength        = 1500
	MaxRecord           = uint(500)
	MaxSeed             = int64(9223372036854775807)
	MinSeed             = int64(100000000000)
	MySQLDateTimeFormat = "2006-01-02 15:04:05"
	lockForUpdate       = "update"
	lockForShare        = "share"
	dbDataStore         = "datastore"
	dbMySQL             = "mysql"
	dbHybrid            = "hybrid"
	modeNormal          = "normal"
	modeTransaction     = "transaction"
	optionTagDatastore  = "datastore" // option tag
	optionTagGoloquent  = "goloquent"
	tagKey              = "__key__"
	tagOmitEmpty        = "omitempty"
	tagNoIndex          = "noindex"
	tagFlatten          = "flatten"
	tagNullable         = "nullable"
	tagUnsigned         = "unsigned" // extra
	tagUnique           = "unique"   // extra
	tagLongText         = "longtext" // extra
)

// SoftDelete :
type SoftDelete struct {
	DeletedDateTime time.Time `goloquent:"DeletedAt" json:"-" xml:"-"`
}

var fieldNameReserved = []string{
	FieldNamePrimaryKey,
	FieldNameParent,
	FieldNameKey,
	FieldNameSoftDelete,
}

// Goloquent Error
var (
	ErrNoSuchEntity         = errors.New("goloquent: entity not found")
	ErrUnsupportDatabase    = errors.New("goloquent: unsupported database type")
	ErrUnsupportDataType    = errors.New("goloquent: unsupported datatype")
	ErrUnsupportFeature     = errors.New("goloquent: database not support this feature")
	ErrInvalidDataTypeModel = errors.New("goloquent: model must be pointer of struct")
	ErrInvalidPrimaryKey    = errors.New("goloquent: invalid primary key")
	ErrMissingPrimaryKey    = errors.New("goloquent: missing primary key")
	ErrParsePrimaryKey      = errors.New("goloquent: unable to parse, invalid key format")
	utf8CharSet             = &CharSet{"utf8", "utf8_unicode_ci"}
	latin2CharSet           = &CharSet{"latin2", "latin2_general_ci"}
	isDebug                 = false
)

var (
	typeOfString          = reflect.TypeOf(string(""))
	typeOfBool            = reflect.TypeOf(bool(false))
	typeOfInt             = reflect.TypeOf(int(0))
	typeOfInt8            = reflect.TypeOf(int8(0))
	typeOfInt16           = reflect.TypeOf(int16(0))
	typeOfInt32           = reflect.TypeOf(int32(0))
	typeOfInt64           = reflect.TypeOf(int64(0))
	typeOfFloat32         = reflect.TypeOf(float32(0))
	typeOfFloat64         = reflect.TypeOf(float64(0))
	typeOfByte            = reflect.TypeOf([]byte(nil))
	typeOfTime            = reflect.TypeOf(time.Time{})
	typeOfDataStoreKey    = reflect.TypeOf(datastore.Key{})
	typeOfPtrDataStoreKey = reflect.TypeOf(&datastore.Key{})
	typeOfGeopoint        = reflect.TypeOf(datastore.GeoPoint{})
	typeOfSoftDelete      = reflect.TypeOf(SoftDelete{})
)
