package model

import (
	"github.com/ungerik/go-start/debug"
	"math"
	"reflect"
	"strconv"
	"unicode"
)

func init() {
	debug.Nop()
}

func Validate(data interface{}, maxDepth int) []*ValidationError {
	errors := []*ValidationError{}
	WalkStructure(data, maxDepth, func(data interface{}, metaData *MetaData) {
		if validator, ok := data.(Validator); ok {
			errors = append(errors, validator.Validate(metaData)...)
		}
	})
	return errors
}

// func IsDefault(data interface{}, maxDepth int) bool {
// 	if field, ok := data.(Field); ok {
// 		return field.IsDefault()
// 	}
// 	v := reflect.ValueOf(data)
// 	for v.Kind() == reflect.Ptr {
// 		v := v.Elem()
// 	}
// 	if v.Kind() == reflect.Struct {

// 	}
// }

type ReportFieldsCallback func(field reflect.Value, metaData *MetaData)

/*
ReportFields reports fields of data via callback.
If data is a struct, array or slice, each field will be reported
recursively until depth is reached.
If depth is 0, no depth limit will be used.
Pointers will be dereferenced without increasing depth or reporting them
until a non pointer value or nil is found.
Cyclic pointer references will lead to an endless loop.
*/
func ReportFields(data interface{}, depth int, callback ReportFieldsCallback) {
	reportFields(reflect.ValueOf(data), nil, depth, callback)
}

func reportFields(v reflect.Value, metaData *MetaData, depth int, callback ReportFieldsCallback) {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Chan, reflect.Invalid:
		return

	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			walkStructure(v.Elem(), metaData, depth, callback)
		}
		return

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			m := &MetaData{
				parent: metaData,
				depth:  metaData.Depth() + 1,
				name:   strconv.Itoa(i),
				index1: i + 1,
			}
			walkStructure(v.Index(i), m, depth, callback)
		}
		return

	case reflect.Struct:
		if metaData.Depth()+1 == depth {
			break
		}
		if v.CanAddr() {
			if _, ok := v.Addr().Interface().(Reference); ok {
				break // Don't go deeper into references
			}
		}
		for i := 0; i < v.NumField(); i++ {
			t := v.Type().Field(i)
			if !unicode.IsUpper(rune(t.Name[0])) {
				continue // Only walk exported fields
			}
			m := metaData
			if !fieldType.Anonymous {
				m = &MetaData{
					parent: metaData,
					pepth:  metaData.Depth() + 1,
					name:   t.Name,
					tag:    t.Tag.Get("gostart"),
				}
			}
			walkStructure(v.Field(i), m, depth, callback)
		}
	}

	// v.Addr() to create a pointer type to enable changing the value and
	// casting to struct types whose methods use pointer to struct
	if v.CanAddr() {
		callback(v.Addr(), metaData)
	}
}
