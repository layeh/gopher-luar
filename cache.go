package luar

import (
	"container/list"
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
	type element struct {
		Type  reflect.Type
		Index []int
	}

	queue := list.New()
	vtype := value.Type()
	queue.PushFront(element{
		Type: vtype,
	})

	for queue.Len() > 0 {
		e := queue.Back()
		elem := e.Value.(element)
		vtype := elem.Type
		if vtype.Kind() == reflect.Ptr {
			vtype = vtype.Elem()
		}
		for i := 0; i < vtype.NumField(); i++ {
			field := vtype.Field(i)
			if field.PkgPath != "" {
				continue
			}
			if tbl.RawGetString(field.Name) != lua.LNil {
				continue
			}
			index := make([]int, len(elem.Index)+1)
			copy(index, elem.Index)
			index[len(elem.Index)] = i

			ud := L.NewUserData()
			ud.Value = index
			tbl.RawSetString(field.Name, ud)
			tbl.RawSetString(getUnexportedName(field.Name), ud)
			if field.Anonymous {
				index := make([]int, len(elem.Index)+1)
				copy(index, elem.Index)
				index[len(elem.Index)] = i
				queue.PushFront(element{
					Type:  field.Type,
					Index: index,
				})
			}
		}

		queue.Remove(e)
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
	case reflect.Array:
		methods := L.NewTable()
		addMethods(L, value, methods)
		mt.RawSetString("methods", methods)

		mt.RawSetString("__index", L.NewFunction(arrayIndex))
		mt.RawSetString("__len", L.NewFunction(arrayLen))
		mt.RawSetString("__eq", L.NewFunction(arrayEq))
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
		mt.RawSetString("methods", methods)

		mt.RawSetString("__index", L.NewFunction(mapIndex))
		mt.RawSetString("__newindex", L.NewFunction(mapNewIndex))
		mt.RawSetString("__len", L.NewFunction(mapLen))
		mt.RawSetString("__call", L.NewFunction(mapCall))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("__eq", L.NewFunction(mapEq))
	case reflect.Ptr:
		ptrMethods := L.NewTable()
		addMethods(L, value, ptrMethods)
		mt.RawSetString("ptr_methods", ptrMethods)
		methods := L.NewTable()
		addMethods(L, value.Elem(), methods)
		mt.RawSetString("methods", methods)
		if value.Elem().Kind() == reflect.Struct {
			fields := L.NewTable()
			addFields(L, value.Elem(), fields)
			mt.RawSetString("fields", fields)
		}

		if value.Elem().Kind() == reflect.Array {
			mt.RawSetString("__index", L.NewFunction(arrayIndex))
		} else {
			mt.RawSetString("__index", L.NewFunction(ptrIndex))
		}
		switch value.Elem().Kind() {
		case reflect.Array:
			mt.RawSetString("__newindex", L.NewFunction(arrayNewIndex))
		case reflect.Struct:
			mt.RawSetString("__newindex", L.NewFunction(structNewIndex))
		}
		mt.RawSetString("__pow", L.NewFunction(ptrPow))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("__unm", L.NewFunction(ptrUnm))
		if value.Elem().Kind() == reflect.Array {
			mt.RawSetString("__len", L.NewFunction(arrayLen))
		}
		mt.RawSetString("__eq", L.NewFunction(ptrEq))
	case reflect.Slice:
		methods := L.NewTable()
		methods.RawSetString("capacity", L.NewFunction(sliceCapacity))
		methods.RawSetString("append", L.NewFunction(sliceAppend))
		addMethods(L, value, methods)
		mt.RawSetString("methods", methods)

		mt.RawSetString("__index", L.NewFunction(sliceIndex))
		mt.RawSetString("__newindex", L.NewFunction(sliceNewIndex))
		mt.RawSetString("__len", L.NewFunction(sliceLen))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
		mt.RawSetString("__eq", L.NewFunction(sliceEq))
	case reflect.Struct:
		methods := L.NewTable()
		addMethods(L, value, methods)
		mt.RawSetString("methods", methods)
		fields := L.NewTable()
		addFields(L, value, fields)
		mt.RawSetString("fields", fields)

		mt.RawSetString("__index", L.NewFunction(structIndex))
		mt.RawSetString("__newindex", L.NewFunction(structNewIndex))
		mt.RawSetString("__tostring", L.NewFunction(allTostring))
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
