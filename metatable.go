package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

type Metatable struct {
	*lua.LTable
}

func MT(L *lua.LState, value interface{}) *Metatable {
	val := reflect.ValueOf(value)
	switch val.Type().Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Struct:
		mt := getMetatable(L, val)
		return &Metatable{mt}
	default:
		return nil
	}
}

func (m *Metatable) Remove(name string) {
	if tbl, ok := m.RawGetString("methods").(*lua.LTable); ok {
		tbl.RawSetString(name, lua.LNil)
	}
	if tbl, ok := m.RawGetString("ptr_methods").(*lua.LTable); ok {
		tbl.RawSetString(name, lua.LNil)
	}
	if tbl, ok := m.RawGetString("fields").(*lua.LTable); ok {
		tbl.RawSetString(name, lua.LNil)
	}
}
