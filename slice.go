package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkSlice(L *lua.LState, idx int) reflect.Value {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Slice {
		L.ArgError(idx, "expecting slice")
	}
	return ref
}

func sliceIndex(L *lua.LState) int {
	ref := checkSlice(L, 1)
	key := L.CheckAny(2)

	switch converted := key.(type) {
	case lua.LNumber:
		index := int(converted)
		if index < 1 || index > ref.Len() {
			L.ArgError(2, "index out of range")
		}
		L.Push(New(L, ref.Index(index-1).Interface()))
	case lua.LString:
		switch string(converted) {
		case "capacity":
			L.Push(L.NewFunction(sliceCapacity))
		case "append":
			L.Push(L.NewFunction(sliceAppend))
		default:
			return 0
		}
	default:
		L.ArgError(2, "must be a number or string")
	}
	return 1
}

func sliceNewIndex(L *lua.LState) int {
	ref := checkSlice(L, 1)
	index := L.CheckInt(2)
	value := L.CheckAny(3)

	if index < 1 || index > ref.Len() {
		L.ArgError(2, "index out of range")
	}
	ref.Index(index - 1).Set(lValueToReflect(value, ref.Type().Elem()))
	return 0
}

func sliceLen(L *lua.LState) int {
	ref := checkSlice(L, 1)
	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func sliceEq(L *lua.LState) int {
	slice1 := checkSlice(L, 1)
	slice2 := checkSlice(L, 2)
	L.Push(lua.LBool(slice1 == slice2))
	return 1
}

// slice methods

func sliceCapacity(L *lua.LState) int {
	ref := checkSlice(L, 1)
	L.Push(lua.LNumber(ref.Cap()))
	return 1
}

func sliceAppend(L *lua.LState) int {
	ref := checkSlice(L, 1)

	hint := ref.Type().Elem()
	values := make([]reflect.Value, L.GetTop()-1)
	for i := 2; i <= L.GetTop(); i++ {
		value := lValueToReflect(L.Get(i), hint)
		if value.Type() != hint {
			L.ArgError(i, "invalid type")
		}
		values[i-2] = value
	}

	newSlice := reflect.Append(ref, values...)
	L.Push(New(L, newSlice.Interface()))
	return 1
}
