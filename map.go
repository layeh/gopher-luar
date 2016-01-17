package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkMap(L *lua.LState, idx int) (reflect.Value, *lua.LTable, bool) {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	isPtr := false
	if ref.Kind() != reflect.Map {
		if ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Map {
			L.ArgError(idx, "expecting map")
		}
		isPtr = true
	}
	return ref, ud.Metatable.(*lua.LTable), isPtr
}

func mapIndex(L *lua.LState) int {
	ref, mt, isPtr := checkMap(L, 1)
	key := L.CheckAny(2)

	if isPtr {
		if lstring, ok := key.(lua.LString); ok {
			if fn := getPtrMethod(string(lstring), mt); fn != nil {
				L.Push(fn)
				return 1
			}
		}
		return 0
	}

	convertedKey := lValueToReflect(key, ref.Type().Key())
	item := ref.MapIndex(convertedKey)
	if !item.IsValid() {
		if lstring, ok := key.(lua.LString); ok {
			if fn := getMethod(string(lstring), mt); fn != nil {
				L.Push(fn)
				return 1
			}
		}
		return 0
	}
	L.Push(New(L, item.Interface()))
	return 1
}

func mapNewIndex(L *lua.LState) int {
	ref, _, isPtr := checkMap(L, 1)
	if isPtr {
		L.RaiseError("invalid operation on map pointer")
	}
	key := L.CheckAny(2)
	value := L.CheckAny(3)

	convertedKey := lValueToReflect(key, ref.Type().Key())
	if convertedKey.Type() != ref.Type().Key() {
		L.ArgError(2, "invalid map key type")
	}
	var convertedValue reflect.Value
	if value != lua.LNil {
		convertedValue = lValueToReflect(value, ref.Type().Elem())
		if convertedValue.Type() != ref.Type().Elem() {
			L.ArgError(3, "invalid map value type")
		}
	}
	ref.SetMapIndex(convertedKey, convertedValue)
	return 0
}

func mapLen(L *lua.LState) int {
	ref, _, isPtr := checkMap(L, 1)
	if isPtr {
		L.RaiseError("invalid operation on map pointer")
	}
	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func mapCall(L *lua.LState) int {
	ref, _, isPtr := checkMap(L, 1)
	if isPtr {
		L.RaiseError("invalid operation on map pointer")
	}
	keys := ref.MapKeys()
	i := 0
	fn := func(L *lua.LState) int {
		if i >= len(keys) {
			return 0
		}
		L.Push(New(L, keys[i].Interface()))
		L.Push(New(L, ref.MapIndex(keys[i]).Interface()))
		i++
		return 2
	}
	L.Push(L.NewFunction(fn))
	return 1
}

func mapEq(L *lua.LState) int {
	map1, _, _ := checkMap(L, 1)
	map2, _, _ := checkMap(L, 2)
	L.Push(lua.LBool(map1 == map2))
	return 1
}
