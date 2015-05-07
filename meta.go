package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

// MetaCall is a struct or struct pointer that defines a fallback action for
// the Lua __call metamethod.
//
// The signature of LuarCall does not matter; it will be converted using the
// standard function conversion rules.
type MetaCall interface {
	LuarCall()
}

func metaCall(L *lua.LState, ref reflect.Value) int {
	refType := ref.Type()
	method, ok := refType.MethodByName("LuarCall")
	if !ok {
		return -1
	}
	return funcEvaluate(L, method.Func)
}

// MetaIndex is a struct or struct pointer that defines a fallback action for
// the Lua __index metamethod.
//
// The signature of LuarIndex does not matter; it will be converted using the
// standard function conversion rules.
type MetaIndex interface {
	LuarIndex()
}

func metaIndex(L *lua.LState, ref reflect.Value) int {
	refType := ref.Type()
	method, ok := refType.MethodByName("LuarIndex")
	if !ok {
		return -1
	}
	return funcEvaluate(L, method.Func)
}

// MetaNewIndex is a struct or struct pointer that defines a fallback action
// for the Lua __newindex metamethod.
//
// The signature of MetaNewIndex does not matter; it will be converted using
// the standard function conversion rules.
type MetaNewIndex interface {
	LuarNewIndex()
}

func metaNewIndex(L *lua.LState, ref reflect.Value) int {
	refType := ref.Type()
	method, ok := refType.MethodByName("LuarNewIndex")
	if !ok {
		return -1
	}
	return funcEvaluate(L, method.Func)
}
