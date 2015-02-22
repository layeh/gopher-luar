package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func getPtrMetaTable(L *lua.LState) lua.LValue {
	key := registryPrefix + "ptr"
	table := L.G.Registry.RawGetH(lua.LString(key))
	if table != lua.LNil {
		return table
	}
	newTable := L.NewTable()
	newTable.RawSetH(lua.LString("__index"), L.NewFunction(ptrIndex))
	newTable.RawSetH(lua.LString("__newindex"), L.NewFunction(ptrNewIndex))
	L.G.Registry.RawSetH(lua.LString(key), newTable)
	return newTable
}

func ptrIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value).Elem()
	switch value.Kind() {
	case reflect.Struct:
		return structIndex(L)
	}
	panic("unsupported pointer type")
	return 0
}

func ptrNewIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value).Elem()
	switch value.Kind() {
	case reflect.Struct:
		return structNewIndex(L)
	}
	panic("unsupported pointer type")
	return 0
}
