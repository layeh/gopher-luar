package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkArray(L *lua.LState, idx int) (reflect.Value, *lua.LTable) {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Array {
		L.ArgError(idx, "expecting array")
	}
	return ref, ud.Metatable.(*lua.LTable)
}

func arrayIndex(L *lua.LState) int {
	ref, mt := checkArray(L, 1)
	key := L.CheckAny(2)

	switch converted := key.(type) {
	case lua.LNumber:
		index := int(converted)
		if index < 1 || index > ref.Len() {
			L.ArgError(2, "index out of range")
		}
		L.Push(New(L, ref.Index(index-1).Interface()))
	case lua.LString:
		if fn := getMethod(converted.String(), mt); fn != nil {
			L.Push(fn)
			return 1
		}
		return 0
	default:
		L.ArgError(2, "must be a number or string")
	}
	return 1
}

func arrayLen(L *lua.LState) int {
	ref, _ := checkArray(L, 1)
	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func arrayEq(L *lua.LState) int {
	array1, _ := checkArray(L, 1)
	array2, _ := checkArray(L, 2)
	L.Push(lua.LBool(array1 == array2))
	return 1
}
