package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkChan(L *lua.LState, idx int) (ref reflect.Value, mt *Metatable, isPtr bool) {
	ud := L.CheckUserData(idx)
	ref = reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Chan {
		if ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Chan {
			L.ArgError(idx, "expecting channel or channel pointer")
		}
		isPtr = true
	}
	mt = &Metatable{LTable: ud.Metatable.(*lua.LTable)}
	return
}

func chanIndex(L *lua.LState) int {
	_, mt, isPtr := checkChan(L, 1)
	key := L.CheckString(2)

	if isPtr {
		if fn := mt.ptrMethod(key); fn != nil {
			L.Push(fn)
			return 1
		}
	}

	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	return 0
}

func chanLen(L *lua.LState) int {
	ref, _, isPtr := checkChan(L, 1)
	if isPtr {
		L.RaiseError("invalid operation on chan pointer")
	}
	L.Push(lua.LNumber(ref.Len()))
	return 1
}

// chan methods

func chanSend(L *lua.LState) int {
	ref, _, _ := checkChan(L, 1)
	value := L.CheckAny(2)
	convertedValue := lValueToReflect(L, value, ref.Type().Elem())
	if convertedValue.Type() != ref.Type().Elem() {
		L.ArgError(2, "incorrect type")
	}
	ref.Send(convertedValue)
	return 0
}

func chanReceive(L *lua.LState) int {
	ref, _, _ := checkChan(L, 1)

	value, ok := ref.Recv()
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LBool(false))
		return 2
	}
	L.Push(New(L, value.Interface()))
	L.Push(lua.LBool(true))
	return 2
}

func chanClose(L *lua.LState) int {
	ref, _, _ := checkChan(L, 1)
	ref.Close()
	return 0
}
