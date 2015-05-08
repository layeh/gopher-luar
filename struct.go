package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkStruct(L *lua.LState, idx int) reflect.Value {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Struct {
		L.ArgError(idx, "expecting struct")
	}
	return ref
}

func structIndex(L *lua.LState) int {
	ref := checkStruct(L, 1)
	refType := ref.Type()

	// Check for method
	key := L.OptString(2, "")
	exKey := getExportedName(key)
	if exKey != "" {
		if method, ok := refType.MethodByName(exKey); ok {
			L.Push(New(L, method.Func.Interface()))
			return 1
		}
	}

	// Check for field
	if field := ref.FieldByName(exKey); field.IsValid() {
		if !field.CanInterface() {
			L.RaiseError("cannot interface field " + exKey)
		}
		L.Push(New(L, field.Interface()))
		return 1
	}

	if ret := metaFunction(L, luarIndexFunc, ref); ret >= 0 {
		return ret
	}
	return 0
}

func structNewIndex(L *lua.LState) int {
	ref := checkStruct(L, 1)
	key := L.OptString(2, "")
	value := L.CheckAny(3)

	if key == "" {
		if ret := metaFunction(L, luarNewIndexFunc, ref); ret >= 0 {
			return ret
		}
		L.TypeError(2, lua.LTString)
		return 0
	}

	exKey := getExportedName(key)

	field := ref.FieldByName(exKey)
	if !field.IsValid() {
		if ret := metaFunction(L, luarNewIndexFunc, ref); ret >= 0 {
			return ret
		}
		L.ArgError(2, "unknown field "+exKey)
	}
	if !field.CanSet() {
		L.ArgError(2, "cannot set field "+exKey)
	}
	field.Set(lValueToReflect(value, field.Type()))

	return 0
}

func structCall(L *lua.LState) int {
	ref := checkStruct(L, 1)
	if ret := metaFunction(L, luarCallFunc, ref); ret >= 0 {
		return ret
	}
	L.RaiseError("attempt to call a non-function object")
	return 0
}
