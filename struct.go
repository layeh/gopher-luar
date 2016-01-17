package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkStruct(L *lua.LState, idx int) (ref reflect.Value, mt *lua.LTable, isPtr bool) {
	ud := L.CheckUserData(idx)
	ref = reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Struct {
		if ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
			L.ArgError(idx, "expecting struct")
		}
		isPtr = true
	}
	mt = ud.Metatable.(*lua.LTable)
	return
}

func structIndex(L *lua.LState) int {
	ref, mt, isPtr := checkStruct(L, 1)
	key := L.CheckString(2)

	if isPtr {
		if fn := getPtrMethod(key, mt); fn != nil {
			L.Push(fn)
			return 1
		}
	}

	if fn := getMethod(key, mt); fn != nil {
		L.Push(fn)
		return 1
	}

	ref = reflect.Indirect(ref)
	index := getFieldIndex(key, mt)
	if index == nil {
		return 0
	}
	field := ref.FieldByIndex(index)
	if !field.CanInterface() {
		L.RaiseError("cannot interface field " + key)
	}
	switch field.Kind() {
	case reflect.Array, reflect.Struct:
		if field.CanAddr() {
			field = field.Addr()
		}
	}
	L.Push(New(L, field.Interface()))
	return 1
}

func structNewIndex(L *lua.LState) int {
	ref, mt, isPtr := checkStruct(L, 1)
	if isPtr {
		ref = ref.Elem()
	}
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
