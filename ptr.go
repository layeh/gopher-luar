package luar

import (
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

func ptrToString(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value)

	str := fmt.Sprintf("userdata: luar: %s %+v (%p)", value.Type(), value.Interface(), ud.Value)
	L.Push(lua.LString(str))
	return 1
}

func ptrIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value).Elem()
	switch value.Kind() {
	case reflect.Struct:
		return structIndex(L)
	}
	L.RaiseError("unsupported pointer type")
	return 0
}

func ptrNewIndex(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := reflect.ValueOf(ud.Value).Elem()
	switch value.Kind() {
	case reflect.Struct:
		return structNewIndex(L)
	}
	L.RaiseError("unsupported pointer type")
	return 0
}
