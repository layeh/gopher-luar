package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func getTypeMetaTable(L *lua.LState) lua.LValue {
	key := registryPrefix + "type"
	table := L.G.Registry.RawGetH(lua.LString(key))
	if table != lua.LNil {
		return table
	}
	newTable := L.NewTable()
	newTable.RawSetH(lua.LString("__call"), L.NewFunction(typeCall))
	L.G.Registry.RawSetH(lua.LString(key), newTable)
	return newTable
}

func typeCall(L *lua.LState) int {
	ud := L.CheckUserData(1)

	refType := ud.Value.(reflect.Type)
	var value reflect.Value
	switch refType.Kind() {
	case reflect.Map:
		value = reflect.MakeMap(refType)
	default:
		value = reflect.New(refType)
	}
	L.Push(New(L, value.Interface()))
	return 1
}
