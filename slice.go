package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkSlice(L *lua.LState, idx int) (ref reflect.Value, mt *Metatable, isPtr bool) {
	ud := L.CheckUserData(idx)
	ref = reflect.ValueOf(ud.Value)
	if ref.Kind() != reflect.Slice {
		if ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Slice {
			L.ArgError(idx, "expecting slice or slice pointer")
		}
		isPtr = true
	}
	mt = &Metatable{LTable: ud.Metatable.(*lua.LTable)}
	return
}

func sliceIndex(L *lua.LState) int {
	ref, mt, isPtr := checkSlice(L, 1)
	key := L.CheckAny(2)

	switch converted := key.(type) {
	case lua.LNumber:
		index := int(converted)
		if index < 1 || index > ref.Len() {
			L.ArgError(2, "index out of range")
		}
		val := ref.Index(index - 1)
		if (val.Kind() == reflect.Struct || val.Kind() == reflect.Array) && val.CanAddr() {
			val = val.Addr()
		}
		L.Push(New(L, val.Interface()))
	case lua.LString:
		if !isPtr {
			if fn := mt.method(string(converted)); fn != nil {
				L.Push(fn)
				return 1
			}
		}
		if fn := mt.ptrMethod(string(converted)); fn != nil {
			L.Push(fn)
			return 1
		}
		return 0
	default:
		L.ArgError(2, "must be a number or string")
	}
	return 1
}

func sliceNewIndex(L *lua.LState) int {
	ref, _, isPtr := checkSlice(L, 1)
	index := L.CheckInt(2)
	value := L.CheckAny(3)

	if isPtr {
		L.RaiseError("invalid operation on slice pointer")
	}

	if index < 1 || index > ref.Len() {
		L.ArgError(2, "index out of range")
	}
	ref.Index(index - 1).Set(lValueToReflect(L, value, ref.Type().Elem(), nil))
	return 0
}

func sliceLen(L *lua.LState) int {
	ref, _, isPtr := checkSlice(L, 1)

	if isPtr {
		L.RaiseError("invalid operation on slice pointer")
	}

	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func sliceCall(L *lua.LState) int {
	ref, _, isPtr := checkSlice(L, 1)
	if isPtr {
		L.RaiseError("invalid operation on slice pointer")
	}

	i := 0
	fn := func(L *lua.LState) int {
		if i >= ref.Len() {
			return 0
		}
		item := ref.Index(i).Interface()
		L.Push(lua.LNumber(i + 1))
		L.Push(New(L, item))
		i++
		return 2
	}

	L.Push(L.NewFunction(fn))
	return 1
}

// slice methods

func sliceCapacity(L *lua.LState) int {
	ref, _, _ := checkSlice(L, 1)
	L.Push(lua.LNumber(ref.Cap()))
	return 1
}

func sliceAppend(L *lua.LState) int {
	ref, _, _ := checkSlice(L, 1)

	hint := ref.Type().Elem()
	values := make([]reflect.Value, L.GetTop()-1)
	for i := 2; i <= L.GetTop(); i++ {
		value := lValueToReflect(L, L.Get(i), hint, nil)
		if value.Type() != hint {
			L.ArgError(i, "invalid type")
		}
		values[i-2] = value
	}

	newSlice := reflect.Append(ref, values...)
	L.Push(New(L, newSlice.Interface()))
	return 1
}
