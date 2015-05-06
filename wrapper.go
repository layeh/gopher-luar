package luar

import (
	"github.com/yuin/gopher-lua"
)

type luaWrapper interface {
	Index(key lua.LValue) (lua.LValue, error)
	NewIndex(key lua.LValue, value lua.LValue) error
	Len() (lua.LValue, error)
	Call(...lua.LValue) ([]lua.LValue, error)
	String() (string, error)
	Equals(v luaWrapper) (lua.LValue, error)

	Unwrap() interface{}
}

func getWrapper(L *lua.LState, idx int) luaWrapper {
	ud := L.CheckUserData(idx)
	wrapper, ok := ud.Value.(luaWrapper)
	if !ok {
		L.RaiseError("invalid datatype")
	}
	return wrapper
}

func wrapperIndex(L *lua.LState) int {
	wrapper := getWrapper(L, 1)
	key := L.CheckAny(2)
	value, err := wrapper.Index(key)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(value)
	return 1
}

func wrapperNewIndex(L *lua.LState) int {
	wrapper := getWrapper(L, 1)
	key := L.CheckAny(2)
	value := L.CheckAny(3)

	if err := wrapper.NewIndex(key, value); err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(value)
	return 1
}

func wrapperLen(L *lua.LState) int {
	wrapper := getWrapper(L, 1)

	value, err := wrapper.Len()
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(value)
	return 1
}

func wrapperCall(L *lua.LState) int {
	wrapper := getWrapper(L, 1)

	args := make([]lua.LValue, L.GetTop()-1)
	for i := 2; i <= L.GetTop(); i++ {
		args[i-2] = L.Get(i)
	}
	values, err := wrapper.Call(args...)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	for _, value := range values {
		L.Push(value)
	}
	return len(values)
}

func wrapperString(L *lua.LState) int {
	wrapper := getWrapper(L, 1)

	str, err := wrapper.String()
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(lua.LString(str))
	return 1
}

func wrapperEq(L *lua.LState) int {
	wrapper1 := getWrapper(L, 1)
	wrapper2 := getWrapper(L, 2)

	value, err := wrapper1.Equals(wrapper2)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(value)
	return 1
}
