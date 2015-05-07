package luar

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/yuin/gopher-lua"
)

type luaSliceWrapper struct {
	L     *lua.LState
	Slice interface{}
}

func (w *luaSliceWrapper) Index(key lua.LValue) (lua.LValue, error) {
	ref := reflect.ValueOf(w.Slice)

	getluaSliceWrapper := func(L *lua.LState) *luaSliceWrapper {
		ud := L.CheckUserData(1)
		w, ok := ud.Value.(*luaSliceWrapper)
		if !ok {
			L.RaiseError("invalid argument (expecting slice)")
		}
		return w
	}

	sliceCapacity := func(L *lua.LState) int {
		w := getluaSliceWrapper(L)
		ref := reflect.ValueOf(w.Slice)
		L.Push(lua.LNumber(ref.Cap()))
		return 1
	}

	sliceAppend := func(L *lua.LState) int {
		w := getluaSliceWrapper(L)
		ref := reflect.ValueOf(w.Slice)

		hint := ref.Type().Elem()
		values := make([]reflect.Value, L.GetTop()-1)
		for i := 2; i <= L.GetTop(); i++ {
			value := lValueToReflect(L.Get(i), hint)
			if value.Type() != hint {
				L.RaiseError("cannot add argument " + strconv.Itoa(i) + " to slice")
			}
			values[i-2] = value
		}

		newSlice := reflect.Append(ref, values...)
		L.Push(New(L, newSlice.Interface()))
		return 1
	}

	switch converted := key.(type) {
	case lua.LNumber:
		intIndex := int(converted)
		if intIndex < 1 || intIndex > ref.Len() {
			return nil, errors.New("index out-of-range")
		}
		return New(w.L, ref.Index(intIndex-1).Interface()), nil
	case lua.LString:
		switch string(converted) {
		case "capacity":
			return w.L.NewFunction(sliceCapacity), nil
		case "append":
			return w.L.NewFunction(sliceAppend), nil
		default:
			return lua.LNil, nil
		}
	}
	return nil, errors.New("index must be a number or a string")
}

func (w *luaSliceWrapper) NewIndex(key lua.LValue, value lua.LValue) error {
	lNumberKey, ok := key.(lua.LNumber)
	if !ok {
		return errors.New("key must be an int")
	}
	ref := reflect.ValueOf(w.Slice)
	index := int(lNumberKey)

	if index < 1 || index > ref.Len() {
		return errors.New("index is out-of-range")
	}
	ref.Index(index - 1).Set(lValueToReflect(value, ref.Type().Elem()))

	return nil
}

func (w *luaSliceWrapper) Len() (lua.LValue, error) {
	ref := reflect.ValueOf(w.Slice)
	return lua.LNumber(ref.Len()), nil
}

func (w *luaSliceWrapper) Call(...lua.LValue) ([]lua.LValue, error) {
	return nil, errors.New("cannot call slice")
}

func (w *luaSliceWrapper) String() (string, error) {
	return getString(w.Slice)
}

func (w *luaSliceWrapper) Equals(other luaWrapper) (lua.LValue, error) {
	v, ok := other.(*luaSliceWrapper)
	if !ok {
		return lua.LFalse, nil
	}
	return lua.LBool(w.Slice == v.Slice), nil
}

func (w *luaSliceWrapper) Unwrap() interface{} {
	return w.Slice
}
