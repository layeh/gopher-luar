package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkPtr(L *lua.LState, idx int) (reflect.Value, *lua.LTable) {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Ptr {
		L.ArgError(idx, "expecting ptr")
	}
	return ref, ud.Metatable.(*lua.LTable)
}

func ptrIndex(L *lua.LState) int {
	ref, mt := checkPtr(L, 1)

	// Check for pointer method
	key := L.OptString(2, "")
	if fn := getPtrMethod(key, mt); fn != nil {
		L.Push(fn)
		return 1
	}

	// Check for method
	if fn := getMethod(key, mt); fn != nil {
		L.Push(fn)
		return 1
	}

	if ref.Elem().Kind() != reflect.Struct {
		return 0
	}

	// Check for struct field
	deref := ref.Elem()
	index := getFieldIndex(key, mt)
	if index == nil {
		return 0
	}
	field := deref.FieldByIndex(index)
	if !field.CanInterface() {
		L.RaiseError("cannot interface field " + key)
	}
	L.Push(New(L, field.Interface()))
	return 1
}

func ptrNewIndex(L *lua.LState) int {
	ref, mt := checkPtr(L, 1)
	deref := ref.Elem()

	if deref.Kind() != reflect.Struct {
		L.RaiseError("cannot new index non-struct pointer")
	}

	key := L.CheckString(2)
	value := L.CheckAny(3)

	index := getFieldIndex(key, mt)
	if index == nil {
		L.RaiseError("unknown field " + key)
	}
	field := deref.FieldByIndex(index)
	if !field.CanSet() {
		L.RaiseError("cannot set field " + key)
	}
	field.Set(lValueToReflect(value, field.Type()))
	return 0
}

func ptrPow(L *lua.LState) int {
	ref, _ := checkPtr(L, 1)
	val := L.CheckAny(2)

	if ref.IsNil() {
		L.RaiseError("cannot dereference nil pointer")
	}
	elem := ref.Elem()
	if !elem.CanSet() {
		L.RaiseError("unable to set pointer value")
	}
	value := lValueToReflect(val, elem.Type())
	elem.Set(value)
	return 1
}

func ptrUnm(L *lua.LState) int {
	ref, _ := checkPtr(L, 1)
	elem := ref.Elem()
	if !elem.CanInterface() {
		L.RaiseError("cannot interface pointer type " + elem.String())
	}
	L.Push(New(L, elem.Interface()))
	return 1
}

func ptrEq(L *lua.LState) int {
	ptr1, _ := checkPtr(L, 1)
	ptr2, _ := checkPtr(L, 2)
	L.Push(lua.LBool(ptr1 == ptr2))
	return 1
}
