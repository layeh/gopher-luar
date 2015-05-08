package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

// Meta can be implemented by a struct or struct pointer. Each method defines
// a fallback action for the corresponding Lua metamethod.
//
// The signature of the methods does not matter; they will be converted using
// the standard function conversion rules. Also, a type is allowed to implement
// only a subset of the interface.
type Meta interface {
	LuarCall(arguments ...interface{}) interface{}
	LuarIndex(key interface{}) interface{}
	LuarNewIndex(key, value interface{})
}

const (
	luarCallFunc     = "LuarCall"
	luarIndexFunc    = "LuarIndex"
	luarNewIndexFunc = "LuarNewIndex"
)

func metaFunction(L *lua.LState, name string, ref reflect.Value) int {
	refType := ref.Type()
	method, ok := refType.MethodByName(name)
	if !ok {
		return -1
	}
	return funcEvaluate(L, method.Func)
}
