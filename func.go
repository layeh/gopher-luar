package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

// LState is an wrapper for gopher-lua's LState. It should be used when you
// wish to have a function/method with the standard "func(*lua.LState) int"
// signature.
type LState struct {
	*lua.LState
}

var refTypeLStatePtr reflect.Type
var refTypeLuaLValueSlice reflect.Type
var refTypeLuaLValue reflect.Type
var refTypeInt reflect.Type

func init() {
	refTypeLStatePtr = reflect.TypeOf(&LState{})
	refTypeLuaLValueSlice = reflect.TypeOf([]lua.LValue{})
	refTypeLuaLValue = reflect.TypeOf((*lua.LValue)(nil)).Elem()
	refTypeInt = reflect.TypeOf(int(0))
}

func funcIsBypass(t reflect.Type) bool {
	if t.NumIn() == 1 && t.NumOut() == 1 && t.In(0) == refTypeLStatePtr && t.Out(0) == refTypeInt {
		return true
	}
	if t.NumIn() == 2 && t.NumOut() == 1 && t.In(1) == refTypeLStatePtr && t.Out(0) == refTypeInt {
		return true
	}
	return false
}

func funcEvaluate(L *lua.LState, fn reflect.Value) int {
	fnType := fn.Type()
	if funcIsBypass(fnType) {
		luarState := LState{L}
		args := make([]reflect.Value, 0, 2)
		if fnType.NumIn() == 2 {
			receiverHint := fnType.In(0)
			receiver := lValueToReflect(L.Get(1), receiverHint)
			if receiver.Type() != receiverHint {
				L.RaiseError("incorrect receiver type")
			}
			args = append(args, receiver)
			L.Remove(1)
		}
		args = append(args, reflect.ValueOf(&luarState))
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
	if len(ret) == 1 && ret[0].Type() == refTypeLuaLValueSlice {
		values := ret[0].Interface().([]lua.LValue)
		for _, value := range values {
			L.Push(value)
		}
		return len(values)
	}
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
