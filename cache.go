package luar

import (
	"reflect"
	"sync"

	"github.com/yuin/gopher-lua"
)

var (
	mu    sync.Mutex
	cache = map[reflect.Type]lua.LValue{}
)

func getMetatable(l *lua.LState, value reflect.Value) lua.LValue {
	mu.Lock()
	defer mu.Unlock()

	vtype := value.Type()
	if v := cache[vtype]; v != nil {
		return v
	}

	tbl := l.NewTable()

	switch vtype.Kind() {
	case reflect.Chan:
		tbl.RawSetString("__index", l.NewFunction(chanIndex))
		tbl.RawSetString("__len", l.NewFunction(chanLen))
		tbl.RawSetString("__tostring", l.NewFunction(allTostring))
		tbl.RawSetString("__eq", l.NewFunction(chanEq))
	case reflect.Map:
		tbl.RawSetString("__index", l.NewFunction(mapIndex))
		tbl.RawSetString("__newindex", l.NewFunction(mapNewIndex))
		tbl.RawSetString("__len", l.NewFunction(mapLen))
		tbl.RawSetString("__call", l.NewFunction(mapCall))
		tbl.RawSetString("__tostring", l.NewFunction(allTostring))
		tbl.RawSetString("__eq", l.NewFunction(mapEq))
	case reflect.Ptr:
		tbl.RawSetString("__index", l.NewFunction(ptrIndex))
		tbl.RawSetString("__newindex", l.NewFunction(ptrNewIndex))
		tbl.RawSetString("__pow", l.NewFunction(ptrPow))
		tbl.RawSetString("__call", l.NewFunction(ptrCall))
		tbl.RawSetString("__tostring", l.NewFunction(allTostring))
		tbl.RawSetString("__unm", l.NewFunction(ptrUnm))
		tbl.RawSetString("__eq", l.NewFunction(ptrEq))
	case reflect.Slice:
		tbl.RawSetString("__index", l.NewFunction(sliceIndex))
		tbl.RawSetString("__newindex", l.NewFunction(sliceNewIndex))
		tbl.RawSetString("__len", l.NewFunction(sliceLen))
		tbl.RawSetString("__tostring", l.NewFunction(allTostring))
		tbl.RawSetString("__eq", l.NewFunction(sliceEq))
	case reflect.Struct:
		tbl.RawSetString("__index", l.NewFunction(structIndex))
		tbl.RawSetString("__newindex", l.NewFunction(structNewIndex))
		tbl.RawSetString("__call", l.NewFunction(structCall))
		tbl.RawSetString("__tostring", l.NewFunction(allTostring))
	}

	cache[vtype] = tbl
	return tbl
}
