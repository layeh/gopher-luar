package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

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
