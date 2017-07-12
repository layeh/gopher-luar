package luar // import "layeh.com/gopher-luar"

import (
	"github.com/yuin/gopher-lua"
)

func chanIndex(L *lua.LState) int {
	_, mt := check(L, 1)
	key := L.CheckString(2)

	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	return 0
}

func chanLen(L *lua.LState) int {
	ref, _ := check(L, 1)

	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func chanEq(L *lua.LState) int {
	ref1, _ := check(L, 1)
	ref2, _ := check(L, 2)

	L.Push(lua.LBool(ref1.Pointer() == ref2.Pointer()))
	return 1
}

func chanCall(L *lua.LState) int {
	ref, _ := check(L, 1)

	switch L.GetTop() {
	// Receive
	case 1:
		value, ok := ref.Recv()
		if ok {
			L.Push(New(L, value.Interface()))
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LNil)
			L.Push(lua.LFalse)
		}
		return 2

	// Send
	case 2:
		value := L.CheckAny(2)

		hint := ref.Type().Elem()
		convertedValue, err := lValueToReflect(L, value, hint, nil)
		if err != nil {
			L.ArgError(2, err.Error())
		}

		ref.Send(convertedValue)
		return 0

	default:
		L.RaiseError("expecting 1 or 2 arguments, got %d", L.GetTop())
		panic("never reaches")
	}
}

func chanUnm(L *lua.LState) int {
	ref, _ := check(L, 1)
	ref.Close()
	return 0
}
