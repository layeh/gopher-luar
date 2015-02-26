package luar

import (
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

func chanToString(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value)

	str := fmt.Sprintf("userdata: luar: %s (%p)", value.Type(), ud.Value)
	L.Push(lua.LString(str))
	return 1
}

func chanSend(L *lua.LState) int {
	ud := L.CheckUserData(1)
	channel := reflect.ValueOf(ud.Value)

	lData := L.Get(2)
	data := lValueToReflect(lData, channel.Type().Elem())
	channel.Send(data)

	return 0
}

func chanReceive(L *lua.LState) int {
	ud := L.CheckUserData(1)
	channel := reflect.ValueOf(ud.Value)

	value, ok := channel.Recv()
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
	ud := L.CheckUserData(1)
	channel := reflect.ValueOf(ud.Value)
	channel.Close()
	return 0
}

func chanIndex(L *lua.LState) int {
	name := L.CheckString(2)

	switch name {
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
