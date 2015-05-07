package luar

import (
	"errors"
	"reflect"

	"github.com/yuin/gopher-lua"
)

type luaChanWrapper struct {
	L    *lua.LState
	Chan interface{}
}

func (w *luaChanWrapper) Index(key lua.LValue) (lua.LValue, error) {
	keyLString, ok := key.(lua.LString)
	if !ok {
		return lua.LNil, nil
	}

	getluaChanWrapper := func(L *lua.LState) *luaChanWrapper {
		ud := L.CheckUserData(1)
		w, ok := ud.Value.(*luaChanWrapper)
		if !ok {
			L.RaiseError("invalid argument (expecting chan)")
		}
		return w
	}

	chanSend := func(L *lua.LState) int {
		w := getluaChanWrapper(L)
		ref := reflect.ValueOf(w.Chan)
		lValue := L.Get(2)
		value := lValueToReflect(lValue, ref.Type().Elem())
		if value.Type() != ref.Type().Elem() {
			L.RaiseError("cannot send given data over the given channel")
			return 0
		}
		ref.Send(value)
		return 0
	}

	chanReceive := func(L *lua.LState) int {
		w := getluaChanWrapper(L)
		ref := reflect.ValueOf(w.Chan)

		value, ok := ref.Recv()
		if !ok {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		L.Push(New(L, value.Interface()))
		L.Push(lua.LBool(true))
		return 2
	}

	chanClose := func(L *lua.LState) int {
		w := getluaChanWrapper(L)
		ref := reflect.ValueOf(w.Chan)
		ref.Close()
		return 0
	}

	switch string(keyLString) {
	case "send":
		return w.L.NewFunction(chanSend), nil
	case "receive":
		return w.L.NewFunction(chanReceive), nil
	case "close":
		return w.L.NewFunction(chanClose), nil
	}
	return lua.LNil, nil
}

func (w *luaChanWrapper) NewIndex(key, value lua.LValue) error {
	return errors.New("cannot set type chan")
}

func (w *luaChanWrapper) Len() (lua.LValue, error) {
	return nil, errors.New("cannot # chan")
}

func (w *luaChanWrapper) Call(...lua.LValue) ([]lua.LValue, error) {
	return nil, errors.New("cannot call chan")
}

func (w *luaChanWrapper) String() (string, error) {
	return getString(w.Chan)
}

func (w *luaChanWrapper) Equals(other luaWrapper) (lua.LValue, error) {
	v, ok := other.(*luaChanWrapper)
	if !ok {
		return lua.LFalse, nil
	}
	return lua.LBool(w.Chan == v.Chan), nil
}

func (w *luaChanWrapper) Unwrap() interface{} {
	return w.Chan
}
