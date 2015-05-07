package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkChan(L *lua.LState, idx int) reflect.Value {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Chan {
		L.ArgError(idx, "expecting chan")
	}
	return ref
}

func chanIndex(L *lua.LState) int {
	_ = checkChan(L, 1)
	key := L.CheckString(2)

	switch key {
	case "send":
		L.Push(L.NewFunction(chanSend))
	case "receive":
		L.Push(L.NewFunction(chanReceive))
	case "close":
		L.Push(L.NewFunction(chanClose))
	default:
		return 0
	}
	return 1
}

func chanLen(L *lua.LState) int {
	ref := checkChan(L, 1)
	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func chanEq(L *lua.LState) int {
	chan1 := checkChan(L, 1)
	chan2 := checkChan(L, 2)
	L.Push(lua.LBool(chan1 == chan2))
	return 1
}

// chan methods

func chanSend(L *lua.LState) int {
	ref := checkChan(L, 1)
	value := L.CheckAny(2)
	convertedValue := lValueToReflect(value, ref.Type().Elem())
	if convertedValue.Type() != ref.Type().Elem() {
		L.ArgError(2, "incorrect type")
	}
	ref.Send(convertedValue)
	return 0
}

func chanReceive(L *lua.LState) int {
	ref := checkChan(L, 1)

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

func chanClose(L *lua.LState) int {
	ref := checkChan(L, 1)
	ref.Close()
	return 0
}
