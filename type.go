package luar

import (
	"errors"
	"reflect"

	"github.com/yuin/gopher-lua"
)

type luaTypeWrapper struct {
	L    *lua.LState
	Type interface{}
}

func (w *luaTypeWrapper) Index(key lua.LValue) (lua.LValue, error) {
	return nil, errors.New("cannot index type")
}

func (w *luaTypeWrapper) NewIndex(key, value lua.LValue) error {
	return errors.New("cannot set type index")
}

func (w *luaTypeWrapper) Len() (lua.LValue, error) {
	return nil, errors.New("cannot # type")
}

func (w *luaTypeWrapper) Call(args ...lua.LValue) ([]lua.LValue, error) {
	ref := w.Type.(reflect.Type)

	var value reflect.Value
	switch ref.Kind() {
	case reflect.Chan:
		var buffer int
		if len(args) >= 1 {
			lNumber, ok := args[0].(lua.LNumber)
			if !ok {
				return nil, errors.New("argument 1 must be an integer")
			}
			buffer = int(lNumber)
		}
		value = reflect.MakeChan(ref, buffer)
	case reflect.Map:
		value = reflect.MakeMap(ref)
	case reflect.Slice:
		var length int
		if len(args) >= 1 {
			lLength, ok := args[0].(lua.LNumber)
			if !ok {
				return nil, errors.New("argument 1 must be an integer")
			}
			length = int(lLength)
		}
		capacity := length
		if len(args) >= 2 {
			lCapacity, ok := args[1].(lua.LNumber)
			if !ok {
				return nil, errors.New("argument 2 must be an integer")
			}
			capacity = int(lCapacity)
		}
		value = reflect.MakeSlice(ref, length, capacity)
	default:
		value = reflect.New(ref)
	}
	return []lua.LValue{New(w.L, value.Interface())}, nil
}

func (w *luaTypeWrapper) String() (string, error) {
	return reflect.ValueOf(w.Type).String(), nil
}

func (w *luaTypeWrapper) Equals(v luaWrapper) (lua.LValue, error) {
	return nil, errors.New("cannot compare type")
}

func (w *luaTypeWrapper) Unwrap() interface{} {
	return w.Type
}
