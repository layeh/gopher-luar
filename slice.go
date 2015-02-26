package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func sliceLen(L *lua.LState) int {
	ud := L.CheckUserData(1)
	slice := reflect.ValueOf(ud.Value)
	L.Push(lua.LNumber(slice.Len()))
	return 1
}

func sliceNewIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	index := L.CheckInt(2)
	value := L.CheckAny(3)

	slice := reflect.ValueOf(ud.Value)
	if index < 1 || index > slice.Len() {
		L.ArgError(2, "index out-of-range")
	}
	slice.Index(index - 1).Set(lValueToReflect(value, slice.Type().Elem()))
	return 0
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
