package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func typeCall(L *lua.LState) int {
	ud := L.CheckUserData(1)

	refType := ud.Value.(reflect.Type)
	var value reflect.Value
	switch refType.Kind() {
	case reflect.Map:
		value = reflect.MakeMap(refType)
	default:
		value = reflect.New(refType)
	}
	L.Push(New(L, value.Interface()))
	return 1
}
