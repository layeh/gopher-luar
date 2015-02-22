package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func getMapMetaTable(L *lua.LState) lua.LValue {
	key := registryPrefix + "map"
	table := L.G.Registry.RawGetH(lua.LString(key))
	if table != lua.LNil {
		return table
	}
	newTable := L.NewTable()
	newTable.RawSetH(lua.LString("__index"), L.NewFunction(mapIndex))
	newTable.RawSetH(lua.LString("__newindex"), L.NewFunction(mapNewIndex))
	newTable.RawSetH(lua.LString("__len"), L.NewFunction(mapLen))
	newTable.RawSetH(lua.LString("__call"), L.NewFunction(mapCall))
	L.G.Registry.RawSetH(lua.LString(key), newTable)
	return newTable
}

func mapLen(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value)
	L.Push(lua.LNumber(value.Len()))
	return 1
}

func mapIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	lKey := L.Get(2)

	value := reflect.ValueOf(ud.Value)
	key := lValueToReflect(lKey, value.Type().Key())
	item := value.MapIndex(key)
	if !item.IsValid() {
		return 0
	}
	L.Push(New(L, item.Interface()))
	return 1
}

func mapNewIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	lKey := L.Get(2)
	lValue := L.Get(3)

	value := reflect.ValueOf(ud.Value)
	key := lValueToReflect(lKey, value.Type().Key())
	mapValue := lValueToReflect(lValue, value.Type().Elem())
	value.SetMapIndex(key, mapValue)
	return 0
}

func mapCall(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value)
	keys := value.MapKeys()
	i := 0
	fn := func(L *lua.LState) int {
		if i >= len(keys) {
			return 0
		}
		L.Push(New(L, keys[i].Interface()))
		L.Push(New(L, value.MapIndex(keys[i]).Interface()))
		i++
		return 2
	}
	L.Push(L.NewFunction(fn))
	return 1
}
