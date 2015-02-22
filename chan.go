package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func getChanMetaTable(L *lua.LState) lua.LValue {
	key := registryPrefix + "chan"
	table := L.G.Registry.RawGetH(lua.LString(key))
	if table != lua.LNil {
		return table
	}
	newTable := L.NewTable()
	newTable.RawSetH(lua.LString("__call"), L.NewFunction(chanCall))
	L.G.Registry.RawSetH(lua.LString(key), newTable)
	return newTable
}

func chanCall(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value)

	switch L.GetTop() {
	case 1: // Receive
		x, ok := value.Recv()
		if !ok {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		L.Push(New(L, x.Interface()))
		L.Push(lua.LBool(true))
		return 2
	case 2: // Send
		xd := L.Get(2)
		x := lValueToReflect(xd, value.Type().Elem())
		value.Send(x)
	default:
		panic("invalid chan call")
	}

	return 0
}
