package luar // import "layeh.com/gopher-luar"

import (
	"container/list"
	"reflect"

	"github.com/yuin/gopher-lua"
)

func addMethods(L *lua.LState, c *Config, vtype reflect.Type, tbl *lua.LTable, ptrReceiver bool) {
	for i := 0; i < vtype.NumMethod(); i++ {
		method := vtype.Method(i)
		if method.PkgPath != "" {
			continue
		}
		namesFn := c.MethodNames
		if namesFn == nil {
			namesFn = defaultMethodNames
		}
		fn := funcWrapper(L, method.Func, ptrReceiver)
		for _, name := range namesFn(vtype, method) {
			tbl.RawSetString(name, fn)
		}
	}
}

func addFields(L *lua.LState, c *Config, vtype reflect.Type, tbl *lua.LTable) {
	type element struct {
		Type  reflect.Type
		Index []int
	}

	queue := list.New()
	queue.PushFront(element{
		Type: vtype,
	})

	namesFn := c.FieldNames
	if namesFn == nil {
		namesFn = defaultFieldNames
	}

	for queue.Len() > 0 {
		e := queue.Back()
		elem := e.Value.(element)
		vtype := elem.Type
	fields:
		for i := 0; i < vtype.NumField(); i++ {
			field := vtype.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				continue
			}
			names := namesFn(vtype, field)
			for _, key := range names {
				if tbl.RawGetString(key) != lua.LNil {
					continue fields
				}
			}
			index := make([]int, len(elem.Index)+1)
			copy(index, elem.Index)
			index[len(elem.Index)] = i

			ud := L.NewUserData()
			ud.Value = index
			for _, key := range names {
				tbl.RawSetString(key, ud)
			}
			if field.Anonymous {
				t := field.Type
				if field.Type.Kind() != reflect.Struct {
					if field.Type.Kind() != reflect.Ptr || field.Type.Elem().Kind() != reflect.Struct {
						continue
					}
					t = field.Type.Elem()
				}
				index := make([]int, len(elem.Index)+1)
				copy(index, elem.Index)
				index[len(elem.Index)] = i
				queue.PushFront(element{
					Type:  t,
					Index: index,
				})
			}
		}

		queue.Remove(e)
	}
}

func getMetatable(L *lua.LState, vtype reflect.Type) *lua.LTable {
	config := GetConfig(L)

	if vtype.Kind() == reflect.Ptr {
		vtype = vtype.Elem()
	}
	if v := config.regular[vtype]; v != nil {
		return v
	}

	var (
		mt         *lua.LTable
		methods    *lua.LTable
		ptrMethods *lua.LTable = L.CreateTable(0, 0)
	)

	switch vtype.Kind() {
	case reflect.Array:
		mt = L.CreateTable(0, 11)
		methods = L.CreateTable(0, 0)

		mt.RawSetString("__index", L.NewFunction(arrayIndex))
		mt.RawSetString("__newindex", L.NewFunction(arrayNewIndex))
		mt.RawSetString("__len", L.NewFunction(arrayLen))
		mt.RawSetString("__call", L.NewFunction(arrayCall))
		mt.RawSetString("__eq", L.NewFunction(arrayEq))
	case reflect.Chan:
		mt = L.CreateTable(0, 9)
		methods = L.CreateTable(0, 3)

		methods.RawSetString("send", L.NewFunction(chanSend))
		methods.RawSetString("receive", L.NewFunction(chanReceive))
		methods.RawSetString("close", L.NewFunction(chanClose))

		mt.RawSetString("__index", L.NewFunction(chanIndex))
		mt.RawSetString("__len", L.NewFunction(chanLen))
		mt.RawSetString("__eq", L.NewFunction(chanEq))
	case reflect.Map:
		mt = L.CreateTable(0, 11)
		methods = L.CreateTable(0, 0)

		mt.RawSetString("__index", L.NewFunction(mapIndex))
		mt.RawSetString("__newindex", L.NewFunction(mapNewIndex))
		mt.RawSetString("__len", L.NewFunction(mapLen))
		mt.RawSetString("__call", L.NewFunction(mapCall))
		mt.RawSetString("__eq", L.NewFunction(mapEq))
	case reflect.Slice:
		mt = L.CreateTable(0, 11)
		methods = L.CreateTable(0, 2)

		methods.RawSetString("capacity", L.NewFunction(sliceCapacity))
		methods.RawSetString("append", L.NewFunction(sliceAppend))

		mt.RawSetString("__index", L.NewFunction(sliceIndex))
		mt.RawSetString("__newindex", L.NewFunction(sliceNewIndex))
		mt.RawSetString("__len", L.NewFunction(sliceLen))
		mt.RawSetString("__call", L.NewFunction(sliceCall))
		mt.RawSetString("__eq", L.NewFunction(sliceEq))
	case reflect.Struct:
		mt = L.CreateTable(0, 10)
		methods = L.CreateTable(0, 0)

		fields := L.NewTable()
		addFields(L, config, vtype, fields)
		mt.RawSetString("fields", fields)

		mt.RawSetString("__index", L.NewFunction(structIndex))
		mt.RawSetString("__newindex", L.NewFunction(structNewIndex))
		mt.RawSetString("__eq", L.NewFunction(structEq))
	default:
		mt = L.CreateTable(0, 8)
		methods = L.CreateTable(0, 0)

		mt.RawSetString("__index", L.NewFunction(ptrIndex))
		mt.RawSetString("__eq", L.NewFunction(ptrEq))
	}

	mt.RawSetString("__tostring", L.NewFunction(tostring))
	mt.RawSetString("__metatable", L.CreateTable(0, 0))
	mt.RawSetString("__pow", L.NewFunction(ptrPow))
	mt.RawSetString("__unm", L.NewFunction(ptrUnm))

	addMethods(L, config, reflect.PtrTo(vtype), ptrMethods, true)
	mt.RawSetString("ptr_methods", ptrMethods)

	addMethods(L, config, vtype, methods, false)
	mt.RawSetString("methods", methods)

	config.regular[vtype] = mt
	return mt
}

func getMetatableFromValue(L *lua.LState, value reflect.Value) *lua.LTable {
	vtype := value.Type()
	return getMetatable(L, vtype)
}

func getTypeMetatable(L *lua.LState, t reflect.Type) *lua.LTable {
	config := GetConfig(L)

	if v := config.types[t]; v != nil {
		return v
	}

	mt := L.NewTable()
	mt.RawSetString("__call", L.NewFunction(typeCall))
	mt.RawSetString("__eq", L.NewFunction(typeEq))

	config.types[t] = mt
	return mt
}
