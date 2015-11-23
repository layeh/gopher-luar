package luar

import (
	"reflect"
	"sync"

	"github.com/yuin/gopher-lua"
)

var (
	mu        sync.Mutex
	cache     = map[reflect.Type]lua.LValue{}
	typeCache = map[reflect.Type]lua.LValue{}
)

func addMethods(L *lua.LState, value reflect.Value, tbl *lua.LTable) {
	vtype := value.Type()
	for i := 0; i < vtype.NumMethod(); i++ {
		method := vtype.Method(i)
		if method.PkgPath != "" {
			continue
		}
		fn := New(L, method.Func.Interface())
		tbl.RawSetString(method.Name, fn)
		tbl.RawSetString(getUnexportedName(method.Name), fn)
	}
}

func addFields(L *lua.LState, value reflect.Value, tbl *lua.LTable) {
	vtype := value.Type()
	for i := 0; i < vtype.NumField(); i++ {
		field := vtype.Field(i)
		if field.PkgPath != "" {
			continue
		}
		ud := L.NewUserData()
		ud.Value = field.Index
		tbl.RawSetString(field.Name, ud)
		tbl.RawSetString(getUnexportedName(field.Name), ud)
	}
}

func getMetatable(L *lua.LState, value reflect.Value) lua.LValue {
	mu.Lock()
	defer mu.Unlock()

	vtype := value.Type()
	if v := cache[vtype]; v != nil {
		return v
	}

	mt := L.NewTable()

	switch vtype.Kind() {
	case reflect.Chan:
		methods := L.NewTable()
		methods.RawSetString("send", L.NewFunction(chanSend))
		methods.RawSetString("receive", L.NewFunction(chanReceive))
		methods.RawSetString("close", L.NewFunction(chanClose))
		addMethods(L, value, methods)

		mt.RawSetString("__index", methods)
		mt.RawSetString("__len", L.NewFunction(chanLen))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("__eq", L.NewFunction(chanEq))
	case reflect.Map:
		methods := L.NewTable()
		addMethods(L, value, methods)

		mt.RawSetString("__index", L.NewFunction(mapIndex))
		mt.RawSetString("__newindex", L.NewFunction(mapNewIndex))
		mt.RawSetString("__len", L.NewFunction(mapLen))
		mt.RawSetString("__call", L.NewFunction(mapCall))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("__eq", L.NewFunction(mapEq))
		mt.RawSetString("methods", methods)
	case reflect.Ptr:
		mt.RawSetString("__index", L.NewFunction(ptrIndex))
		mt.RawSetString("__newindex", L.NewFunction(ptrNewIndex))
		mt.RawSetString("__pow", L.NewFunction(ptrPow))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("__unm", L.NewFunction(ptrUnm))
		mt.RawSetString("__eq", L.NewFunction(ptrEq))
	case reflect.Slice:
		methods := L.NewTable()
		methods.RawSetString("capacity", L.NewFunction(sliceCapacity))
		methods.RawSetString("append", L.NewFunction(sliceAppend))
		addMethods(L, value, methods)

		mt.RawSetString("__index", L.NewFunction(sliceIndex))
		mt.RawSetString("__newindex", L.NewFunction(sliceNewIndex))
		mt.RawSetString("__len", L.NewFunction(sliceLen))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("__eq", L.NewFunction(sliceEq))
		mt.RawSetString("methods", methods)
	case reflect.Struct:
		methods := L.NewTable()
		addMethods(L, value, methods)
		mt.RawSetString("__index", L.NewFunction(structIndex))
		mt.RawSetString("__newindex", L.NewFunction(structNewIndex))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("methods", methods)
	}

	cache[vtype] = mt
	return mt
}

func getTypeMetatable(L *lua.LState, t reflect.Type) lua.LValue {
	mu.Lock()
	defer mu.Unlock()

	if v := typeCache[t]; v != nil {
		return v
	}

	mt := L.NewTable()
	mt.RawSetString("__call", L.NewFunction(typeCall))
	mt.RawSetString("__tostring", L.NewFunction(allTostring))
	mt.RawSetString("__eq", L.NewFunction(typeEq))

	typeCache[t] = mt
	return mt
}
