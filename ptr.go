package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkPtr(L *lua.LState, idx int) reflect.Value {
	ud := L.CheckUserData(idx)
	ref := reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Ptr {
		L.ArgError(idx, "expecting ptr")
	}
	return ref
}

func ptrIndex(L *lua.LState) int {
	ref := checkPtr(L, 1)
	deref := ref.Elem()
	if deref.Kind() != reflect.Struct {
		L.RaiseError("cannot index non-struct pointer")
	}
	refType := ref.Type()

	// Check for method
	key := L.OptString(2, "")
	exKey := getExportedName(key)
	if key != "" {
		if method, ok := refType.MethodByName(exKey); ok {
			L.Push(New(L, method.Func.Interface()))
			return 1
		}
	}

	// Check for field
	if field := deref.FieldByName(exKey); field.IsValid() {
		if !field.CanInterface() {
			L.RaiseError("cannot interface field " + exKey)
		}
		if val := New(L, field.Interface()); val != nil {
			L.Push(val)
			return 1
		}
		L.RaiseError("could not convert field " + exKey)
	}

	if ret := metaFunction(L, luarIndexFunc, ref); ret >= 0 {
		return ret
	}
	return 0
}

func ptrNewIndex(L *lua.LState) int {
	ref := checkPtr(L, 1)
	deref := ref.Elem()

	if deref.Kind() != reflect.Struct {
		L.RaiseError("cannot new index non-struct pointer")
	}

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

	field := deref.FieldByName(exKey)
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

func ptrPow(L *lua.LState) int {
	ref := checkPtr(L, 1)
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

func ptrLen(L *lua.LState) int {
	ref := checkPtr(L, 1)
	L.Push(lua.LBool(!ref.IsNil()))
	return 1
}

func ptrCall(L *lua.LState) int {
	ref := checkPtr(L, 1)
	if ret := metaFunction(L, luarCallFunc, ref); ret >= 0 {
		return ret
	}
	L.RaiseError("attempt to call a non-function object")
	return 0
}

func ptrUnm(L *lua.LState) int {
	ref := checkPtr(L, 1)
	elem := ref.Elem()
	if !elem.CanInterface() {
		L.RaiseError("cannot interface pointer type " + elem.String())
	}
	L.Push(New(L, elem.Interface()))
	return 1
}

func ptrEq(L *lua.LState) int {
	ptr1 := checkPtr(L, 1)
	ptr2 := checkPtr(L, 2)
	L.Push(lua.LBool(ptr1 == ptr2))
	return 1
}
