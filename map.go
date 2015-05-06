package luar

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

type luaMapWrapper struct {
	L   *lua.LState
	Map interface{}
}

func (w *luaMapWrapper) Index(key lua.LValue) (lua.LValue, error) {
	ref := reflect.ValueOf(w.Map)

	convertedKey := lValueToReflect(key, ref.Type().Key())
	item := ref.MapIndex(convertedKey)
	if !item.IsValid() {
		return lua.LNil, nil
	}
	return New(w.L, item.Interface()), nil
}

func (w *luaMapWrapper) NewIndex(key lua.LValue, value lua.LValue) error {
	ref := reflect.ValueOf(w.Map)

	convertedKey := lValueToReflect(key, ref.Type().Key())
	if convertedKey.Type() != ref.Type().Key() {
		return errors.New("invalid map key type")
	}
	var convertedValue reflect.Value
	if value != lua.LNil {
		convertedValue = lValueToReflect(value, ref.Type().Elem())
		if convertedValue.Type() != ref.Type().Elem() {
			return errors.New("invalid map value type")
		}
	}
	ref.SetMapIndex(convertedKey, convertedValue)
	return nil
}

func (w *luaMapWrapper) Len() (lua.LValue, error) {
	ref := reflect.ValueOf(w.Map)
	return lua.LNumber(ref.Len()), nil
}

func (w *luaMapWrapper) Call(...lua.LValue) ([]lua.LValue, error) {
	ref := reflect.ValueOf(w.Map)
	keys := ref.MapKeys()
	i := 0
	fn := func(L *lua.LState) int {
		if i >= len(keys) {
			return 0
		}
		L.Push(New(L, keys[i].Interface()))
		L.Push(New(L, ref.MapIndex(keys[i]).Interface()))
		i++
		return 2
	}
	return []lua.LValue{w.L.NewFunction(fn)}, nil
}

func (w *luaMapWrapper) String() (string, error) {
	if stringer, ok := w.Map.(fmt.Stringer); ok {
		return stringer.String(), nil
	}
	return reflect.ValueOf(w.Map).String(), nil
}

func (w *luaMapWrapper) Equals(other luaWrapper) (lua.LValue, error) {
	v, ok := other.(*luaMapWrapper)
	if !ok {
		return lua.LFalse, nil
	}
	return lua.LBool(w.Map == v.Map), nil
}

func (w *luaMapWrapper) Unwrap() interface{} {
	return w.Map
}
