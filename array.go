package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkArray(L *lua.LState, idx int) (reflect.Value, *lua.LTable) {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Array && (ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Array) {
		L.ArgError(idx, "expecting array")
	}
	return ref, ud.Metatable.(*lua.LTable)
}

func arrayIndex(L *lua.LState) int {
	ref, mt := checkArray(L, 1)
	ref = reflect.Indirect(ref)
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

func arrayNewIndex(L *lua.LState) int {
	ref, _ := checkArray(L, 1)
	deref := ref.Elem()

	index := L.CheckInt(2)
	value := L.CheckAny(3)
	if index < 1 || index > deref.Len() {
		L.ArgError(2, "index out of range")
	}
	deref.Index(index - 1).Set(lValueToReflect(value, deref.Type().Elem()))
	return 0
}

func arrayLen(L *lua.LState) int {
	ref, _ := checkArray(L, 1)
	ref = reflect.Indirect(ref)
	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func arrayEq(L *lua.LState) int {
	array1, _ := checkArray(L, 1)
	array2, _ := checkArray(L, 2)
	L.Push(lua.LBool(array1 == array2))
	return 1
}
