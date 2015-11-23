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

	exKey := getExportedName(key)

	// Check for field
	if field := ref.FieldByName(exKey); field.IsValid() {
		if !field.CanInterface() {
			L.RaiseError("cannot interface field " + exKey)
		}
		L.Push(New(L, field.Interface()))
		return 1
	}
	return 0
}

func structNewIndex(L *lua.LState) int {
	ref, _ := checkStruct(L, 1)
	key := L.CheckString(2)
	value := L.CheckAny(3)

	exKey := getExportedName(key)

	field := ref.FieldByName(exKey)
	if !field.IsValid() {
		L.ArgError(2, "unknown field "+exKey)
	}
	if !field.CanSet() {
		L.ArgError(2, "cannot set field "+exKey)
	}
	field.Set(lValueToReflect(value, field.Type()))

	return 0
}
