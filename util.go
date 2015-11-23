package luar

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/gopher-lua"
)

func allTostring(L *lua.LState) int {
	ud := L.CheckUserData(1)
	value := ud.Value
	if stringer, ok := value.(fmt.Stringer); ok {
		L.Push(lua.LString(stringer.String()))
	} else {
		L.Push(lua.LString(fmt.Sprintf("userdata (luar): %p", ud)))
	}
	return 1
}

func getExportedName(name string) string {
	buf := []byte(name)
	first, n := utf8.DecodeRune(buf)
	if n == 0 {
		return name
	}
	return string(unicode.ToUpper(first)) + string(buf[n:])
}

func getUnexportedName(name string) string {
	buf := []byte(name)
	first, n := utf8.DecodeRune(buf)
	if n == 0 {
		return name
	}
	return string(unicode.ToLower(first)) + string(buf[n:])
}

func getMethod(key string, mt *lua.LTable) lua.LValue {
	methods := mt.RawGetString("methods").(*lua.LTable)
	if fn := methods.RawGetString(key); fn != lua.LNil {
		return fn
	}
	return nil
}
