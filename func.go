package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func getLuaFuncWrapper(L *lua.LState, fn reflect.Value) *lua.LFunction {
	wrapper := func(L *lua.LState) int {
		fnType := fn.Type()
		top := L.GetTop()
		expected := fnType.NumIn()
		if top != expected {
			panic("invalid number of function arguments")
		}
		args := make([]reflect.Value, top)
		for i := 0; i < L.GetTop(); i++ {
			args[i] = lValueToReflect(L.Get(i+1), fnType.In(i))
		}
		ret := fn.Call(args)
		for _, val := range ret {
			L.Push(New(L, val.Interface()))
		}
		return len(ret)
	}
	return L.NewFunction(wrapper)
}
