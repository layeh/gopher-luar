package luar

import (
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

func typeToString(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := ud.Value.(reflect.Type)

	str := fmt.Sprintf("userdata: luar: type %s", value)
	L.Push(lua.LString(str))
	return 1
}

func typeCall(L *lua.LState) int {
	ud := L.CheckUserData(1)

	refType := ud.Value.(reflect.Type)
	var value reflect.Value
	switch refType.Kind() {
	case reflect.Chan:
		buffer := L.OptInt(2, 0)
		value = reflect.MakeChan(refType, buffer)
	case reflect.Map:
		value = reflect.MakeMap(refType)
	case reflect.Slice:
		length := L.CheckInt(2)
		capacity := L.OptInt(3, length)
		value = reflect.MakeSlice(refType, length, capacity)
	default:
		value = reflect.New(refType)
	}
	L.Push(New(L, value.Interface()))
	return 1
}
