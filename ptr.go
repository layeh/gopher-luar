package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func checkPtr(L *lua.LState, idx int) (ref reflect.Value, opts ReflectOptions, mt *Metatable) {
	ud := L.CheckUserData(idx)
	refIface, ok := ud.Value.(*ReflectedInterface)
	if !ok {
		L.RaiseError("unexpected userdata value")
	}

	ref = reflect.ValueOf(refIface.Interface)
	opts = refIface.Options

	kind := reflect.Ptr
	if ref.Kind() != kind {
		L.ArgError(idx, "expecting "+kind.String())
	}
	mt = &Metatable{LTable: ud.Metatable.(*lua.LTable)}
	return
}

//func checkPtr(L *lua.LState, idx int) (ref reflect.Value, opts ReflectOptions, mt *Metatable) {
//	ud := L.CheckUserData(idx)
//
//	if refIface, ok := ud.Value.(*ReflectedInterface); ok {
//		ref = reflect.ValueOf(refIface.Interface)
//		opts = refIface.Options
//	} else {
//		ref = reflect.ValueOf(ud.Value)
//	}
//
//	kind := reflect.Ptr
//	if ref.Kind() != kind {
//		L.ArgError(idx, "expecting "+kind.String())
//	}
//	mt = &Metatable{LTable: ud.Metatable.(*lua.LTable)}
//	return
//}

func ptrIndex(L *lua.LState) int {
	_, _, mt := checkPtr(L, 1)
	key := L.CheckString(2)

	if fn := mt.ptrMethod(key); fn != nil {
		L.Push(fn)
		return 1
	}

	if fn := mt.method(key); fn != nil {
		L.Push(fn)
		return 1
	}

	return 0
}

func ptrPow(L *lua.LState) int {
	ref, opts, _ := checkPtr(L, 1)

	if opts.Immutable {
		L.RaiseError("invalid operation for immutable pointer")
	}

	val := L.CheckAny(2)

	if ref.IsNil() {
		L.RaiseError("cannot dereference nil pointer")
	}
	elem := ref.Elem()
	if !elem.CanSet() {
		L.RaiseError("unable to set pointer value")
	}
	value := lValueToReflect(L, val, elem.Type(), nil)
	elem.Set(value)
	return 1
}

func ptrUnm(L *lua.LState) int {
	ref, _, _ := checkPtr(L, 1)
	elem := ref.Elem()
	if !elem.CanInterface() {
		L.RaiseError("cannot interface pointer type " + elem.String())
	}
	L.Push(New(L, elem.Interface()))
	return 1
}
