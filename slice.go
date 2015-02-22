package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func getSliceMetaTable(L *lua.LState) lua.LValue {
	key := registryPrefix + "slice"
	table := L.G.Registry.RawGetH(lua.LString(key))
	if table != lua.LNil {
		return table
	}
	newTable := L.NewTable()
	newTable.RawSetH(lua.LString("__index"), L.NewFunction(sliceIndex))
	newTable.RawSetH(lua.LString("__len"), L.NewFunction(sliceLen))
	L.G.Registry.RawSetH(lua.LString(key), newTable)
	return newTable
}

func sliceLen(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value)
	L.Push(lua.LNumber(value.Len()))
	return 1
}

func sliceIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	lIndex := L.Get(2)

	value := reflect.ValueOf(ud.Value)
	fIndex, ok := lIndex.(lua.LNumber)
	if !ok {
		return 0
	}
	index := int(fIndex)
	if index < 1 || index > value.Len() {
		return 0
	}
	L.Push(New(L, value.Index(index-1).Interface()))
	return 1
}
