package luar // import "layeh.com/gopher-luar"

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func structIndex(L *lua.LState) int {
	ref, mt, isPtr := check(L, 1, reflect.Struct)
	key := L.CheckString(2)

	if !isPtr {
		if fn := mt.method(key); fn != nil {
			L.Push(fn)
			return 1
		}
	}

	if fn := mt.ptrMethod(key); fn != nil {
		L.Push(fn)
		return 1
	}

	ref = reflect.Indirect(ref)
	index := mt.fieldIndex(key)
	if index == nil {
		return 0
	}
	field := ref.FieldByIndex(index)
	if !field.CanInterface() {
		L.RaiseError("cannot interface field " + key)
	}

	if (field.Kind() == reflect.Struct || field.Kind() == reflect.Array) && field.CanAddr() {
		field = field.Addr()
	}
	L.Push(New(L, field.Interface()))
	return 1
}

func structNewIndex(L *lua.LState) int {
	ref, mt, isPtr := check(L, 1, reflect.Struct)
	if isPtr {
		ref = ref.Elem()
	}
	key := L.CheckString(2)
	value := L.CheckAny(3)

	index := mt.fieldIndex(key)
	if index == nil {
		L.RaiseError("unknown field " + key)
	}
	field := ref.FieldByIndex(index)
	if !field.CanSet() {
		L.RaiseError("cannot set field " + key)
	}
	val := lValueToReflect(L, value, field.Type(), nil)
	if !val.IsValid() {
		raiseInvalidArg(L, 2, value, field.Type())
	}
	field.Set(val)
	return 0
}

func structEq(L *lua.LState) int {
	ref1, _, isPtr1 := check(L, 1, reflect.Struct)
	ref2, _, isPtr2 := check(L, 2, reflect.Struct)

	if (isPtr1 && isPtr2) || (!isPtr1 && !isPtr2) {
		L.Push(lua.LBool(ref1.Interface() == ref2.Interface()))
		return 1
	}

	L.RaiseError("invalid operation == on mixed struct value and pointer")
	return 0 // never reaches
}
