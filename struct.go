package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkStruct(L *lua.LState, idx int) (reflect.Value, *lua.LTable) {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Struct {
		L.ArgError(idx, "expecting struct")
	}
	return ref, ud.Metatable.(*lua.LTable)
}

func structIndex(L *lua.LState) int {
	ref, mt := checkStruct(L, 1)
	key := L.CheckString(2)

	// Check for method
	if fn := getMethod(key, mt); fn != nil {
		L.Push(fn)
		return 1
	}

	// Check for field
	index := getFieldIndex(key, mt)
	if index == nil {
		return 0
	}
	field := ref.FieldByIndex(index)
	if !field.CanInterface() {
		L.RaiseError("cannot interface field " + key)
	}
	L.Push(New(L, field.Interface()))
	return 1
}

func structNewIndex(L *lua.LState) int {
	ref, mt := checkStruct(L, 1)
	key := L.CheckString(2)
	value := L.CheckAny(3)

	index := getFieldIndex(key, mt)
	if index == nil {
		L.RaiseError("unknown field " + key)
	}
	field := ref.FieldByIndex(index)
	if !field.CanSet() {
		L.RaiseError("cannot set field " + key)
	}
	field.Set(lValueToReflect(value, field.Type()))
	return 0
}
