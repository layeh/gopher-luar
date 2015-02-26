package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func structMethod(L *lua.LState) int {
	ud := L.CheckUserData(1)
	name := L.CheckString(lua.UpvalueIndex(1))

	L.Remove(1)

	value := reflect.ValueOf(ud.Value)
	if value.Kind() == reflect.Ptr {
		method := value.MethodByName(name)
		return funcEvaluate(L, method)
	}

	method := value.MethodByName(name)
	return funcEvaluate(L, method)
}

func structIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	name := L.CheckString(2)

	value := reflect.ValueOf(ud.Value)
	if value.Kind() == reflect.Ptr {
		if method := value.MethodByName(name); method.IsValid() {
			L.Push(L.NewClosure(structMethod, lua.LString(name)))
			return 1
		}
		value = value.Elem()
	}

	if method := value.MethodByName(name); method.IsValid() {
		L.Push(L.NewClosure(structMethod, lua.LString(name)))
		return 1
	}

	field := value.FieldByName(name)
	if field.IsValid() {
		if val := New(L, field.Interface()); val != nil {
			L.Push(val)
			return 1
		}
	}

	return 0
}

func structNewIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	name := L.CheckString(2)
	lValue := L.Get(3)

	value := reflect.ValueOf(ud.Value)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	field := value.FieldByName(name)
	field.Set(lValueToReflect(lValue, field.Type()))
	return 0
}
