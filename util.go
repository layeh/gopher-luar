package luar

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/gopher-lua"
)

func tostring(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := ud.Value
	if stringer, ok := value.(fmt.Stringer); ok {
		L.Push(lua.LString(stringer.String()))
	} else {
		L.Push(lua.LString(fmt.Sprintf("userdata (luar): %p", ud)))
	}
	return 1
}

func eq(L *lua.LState) int {
	ud1 := L.CheckUserData(1).Value
	ud2 := L.CheckUserData(2).Value
	L.Push(lua.LBool(reflect.DeepEqual(ud1, ud2)))
	return 1
}

func getUnexportedName(name string) string {
	first, n := utf8.DecodeRuneInString(name)
	if n == 0 {
		return name
	}
	return string(unicode.ToLower(first)) + name[n:]
}
