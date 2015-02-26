package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func sliceCapacity(L *lua.LState) int {
	ud := L.CheckUserData(1)
	slice := reflect.ValueOf(ud.Value)
	L.Push(lua.LNumber(slice.Cap()))
	return 1
}

func sliceAppend(L *lua.LState) int {
	ud := L.CheckUserData(1)
	slice := reflect.ValueOf(ud.Value)

	hint := slice.Type().Elem()
	values := make([]reflect.Value, L.GetTop()-1)
	for i := 2; i <= L.GetTop(); i++ {
		values[i-2] = lValueToReflect(L.Get(i), hint)
	}

	newSlice := reflect.Append(slice, values...)
	L.Push(New(L, newSlice.Interface()))
	return 1
}

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
	index := L.Get(2)

	slice := reflect.ValueOf(ud.Value)
	switch index := index.(type) {
	case lua.LNumber:
		intIndex := int(index)
		if intIndex < 1 || intIndex > slice.Len() {
			L.ArgError(2, "index out-of-range")
		}
		L.Push(New(L, slice.Index(intIndex-1).Interface()))
	case lua.LString:
		switch string(index) {
		case "capacity":
			L.Push(L.NewFunction(sliceCapacity))
		case "append":
			L.Push(L.NewFunction(sliceAppend))
		default:
			return 0
		}
	default:
		L.ArgError(2, "index must be a number or a string")
	}
	return 1
}
