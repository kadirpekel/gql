package gql

import (
	"context"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
)

var (
	ContextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	InfoType    = reflect.TypeOf((*graphql.ResolveInfo)(nil)).Elem()
	ErrorType   = reflect.TypeOf((*error)(nil)).Elem()
)

type ArgInfo struct {
	Type     reflect.Type
	RealType reflect.Type
	Index    int
	IsPtr    bool
	IsSlice  bool
}

func NewArgInfo(argType reflect.Type, index int) *ArgInfo {
	realType := argType
	isPtr := argType.Kind() == reflect.Ptr
	isSlice := argType.Kind() == reflect.Slice
	if isPtr || isSlice {
		realType = argType.Elem()
	}
	return &ArgInfo{
		Index:    index,
		Type:     argType,
		IsPtr:    isPtr,
		RealType: realType,
		IsSlice:  isSlice,
	}
}

func (a *ArgInfo) ValueFromMap(m interface{}) (reflect.Value, error) {
	obj := reflect.New(a.RealType).Interface()
	err := mapstructure.Decode(m, obj)
	if err != nil {
		return reflect.Value{}, err
	}
	if a.IsPtr {
		return reflect.ValueOf(obj), nil
	}
	return reflect.ValueOf(obj).Elem(), nil
}

func (a *ArgInfo) ValueFromSlice(value interface{}) (reflect.Value, error) {
	length := reflect.ValueOf(value).Len()
	slice := reflect.MakeSlice(a.Type, length, length)
	for i := 0; i < length; i++ {
		elem, err := a.ValueFrom(reflect.ValueOf(value).Index(i).Interface())
		if err != nil {
			return reflect.Value{}, err
		}
		slice.Index(i).Set(elem.Elem())
	}
	return slice, nil
}

func (a *ArgInfo) ValueFrom(value interface{}) (reflect.Value, error) {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		if a.IsPtr {
			return reflect.ValueOf(value), nil
		}
		return reflect.ValueOf(value).Elem(), nil
	} else if reflect.TypeOf(value).Kind() == reflect.Map {
		return a.ValueFromMap(value.(map[string]interface{}))
	} else if reflect.TypeOf(value).Kind() == reflect.Slice {
		return a.ValueFromSlice(value)
	} else {
		if a.IsPtr {
			ptr := reflect.New(a.RealType)
			ptr.Elem().Set(reflect.ValueOf(value))
			return ptr, nil
		}
		return reflect.ValueOf(value), nil
	}
}
