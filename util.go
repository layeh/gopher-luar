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

func getUnexportedName(name string) string {
	first, n := utf8.DecodeRuneInString(name)
	if n == 0 {
		return name
	}
	return string(unicode.ToLower(first)) + name[n:]
}

func getMethod(key string, mt *lua.LTable) lua.LValue {
	methods := mt.RawGetString("methods").(*lua.LTable)
	if fn := methods.RawGetString(key); fn != lua.LNil {
		return fn
	}
	return nil
}

func getPtrMethod(key string, mt *lua.LTable) lua.LValue {
	methods := mt.RawGetString("ptr_methods").(*lua.LTable)
	if fn := methods.RawGetString(key); fn != lua.LNil {
		return fn
	}
	return nil
}

func getFieldIndex(key string, mt *lua.LTable) []int {
	fields := mt.RawGetString("fields").(*lua.LTable)
	if index := fields.RawGetString(key); index != lua.LNil {
		return index.(*lua.LUserData).Value.([]int)
	}
	return nil
}
