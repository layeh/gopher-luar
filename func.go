package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

// LState is an wrapper for gopher-lua's LState. It should be used when you
// wish to have a function/method with the standard "func(*lua.LState) int"
// signature.
//
// Example:
//  fn := func(L *luar.LState) int {
//    fmt.Printf("Got %d arguments\n", L.GetTop())
//    return 0
//  }
//
//  L.SetGlobal("fn", luar.New(L, fn))
//  --
//  fn(1, 2, 3, 4, 5) // prints "Got 5 arguments"
type LState struct {
	*lua.LState
}

var lStatePtrType reflect.Type
var intType reflect.Type

func init() {
	lStatePtrType = reflect.TypeOf(&LState{})
	intType = reflect.TypeOf(int(0))
}

func funcIsBypass(t reflect.Type) bool {
	return t.NumIn() == 1 && t.NumOut() == 1 && t.In(0) == lStatePtrType && t.Out(0) == intType
}

func funcEvaluate(L *lua.LState, fn reflect.Value) int {
	fnType := fn.Type()
	if funcIsBypass(fnType) {
		luarState := LState{L}
		args := []reflect.Value{
			reflect.ValueOf(&luarState),
		}
		return fn.Call(args)[0].Interface().(int)
	}

	top := L.GetTop()
	expected := fnType.NumIn()
	variadic := fnType.IsVariadic()
	if !variadic && top != expected {
		L.RaiseError("invalid number of function argument (%d expected, got %d)", expected, top)
	}
	if variadic && top < expected-1 {
		L.RaiseError("invalid number of function argument (%d or more expected, got %d)", expected-1, top)
	}
	args := make([]reflect.Value, top)
	for i := 0; i < L.GetTop(); i++ {
		var hint reflect.Type
		if variadic && i >= expected-1 {
			hint = fnType.In(expected - 1).Elem()
		} else {
			hint = fnType.In(i)
		}
		args[i] = lValueToReflect(L.Get(i+1), hint)
	}
	ret := fn.Call(args)
	for _, val := range ret {
		L.Push(New(L, val.Interface()))
	}
	return len(ret)
}

func funcWrapper(L *lua.LState, fn reflect.Value) *lua.LFunction {
	wrapper := func(L *lua.LState) int {
		return funcEvaluate(L, fn)
	}
	return L.NewFunction(wrapper)
}
