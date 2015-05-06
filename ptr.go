package luar

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

type luaPointerWrapper struct {
	L   *lua.LState
	Ptr interface{}
}

func (w *luaPointerWrapper) Index(key lua.LValue) (lua.LValue, error) {
	ref := reflect.ValueOf(w.Ptr)
	refType := ref.Type()

	// Check for method
	keyLString, ok := key.(lua.LString)
	if ok {
		if method, ok := refType.MethodByName(string(keyLString)); ok {
			return New(w.L, method.Func.Interface()), nil
		}
	}

	// Check for field
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		return nil, errors.New("cannot index non-struct pointer type")
	}

	if field := ref.FieldByName(string(keyLString)); field.IsValid() {
		if !field.CanInterface() {
			return nil, errors.New("cannot interface field " + string(keyLString))
		}
		if val := New(w.L, field.Interface()); val != nil {
			return val, nil
		}
	}

	return lua.LNil, nil
}

func (w *luaPointerWrapper) NewIndex(key lua.LValue, value lua.LValue) error {
	ref := reflect.ValueOf(w.Ptr).Elem()

	if ref.Kind() != reflect.Struct {
		return errors.New("cannot new index non-struct pointer type")
	}

	keyLString, ok := key.(lua.LString)
	if !ok {
		return errors.New("invalid non-string key")
	}

	keyString := string(keyLString)
	field := ref.FieldByName(keyString)
	if !field.IsValid() {
		return errors.New("unknown field " + keyString)
	}
	if !field.CanSet() {
		return errors.New("cannot set field " + keyString)
	}
	field.Set(lValueToReflect(value, field.Type()))
	return nil
}

func (w *luaPointerWrapper) Len() (lua.LValue, error) {
	return nil, errors.New("cannot # ptr")
}

func (w *luaPointerWrapper) Call(...lua.LValue) ([]lua.LValue, error) {
	return nil, errors.New("cannot call ptr")
}

func (w *luaPointerWrapper) String() (string, error) {
	if stringer, ok := w.Ptr.(fmt.Stringer); ok {
		return stringer.String(), nil
	}
	return reflect.ValueOf(w.Ptr).String(), nil
}

func (w *luaPointerWrapper) Equals(other luaWrapper) (lua.LValue, error) {
	v, ok := other.(*luaPointerWrapper)
	if !ok {
		return lua.LFalse, nil
	}
	return lua.LBool(w.Ptr == v.Ptr), nil
}

func (w *luaPointerWrapper) Unwrap() interface{} {
	return w.Ptr
}
