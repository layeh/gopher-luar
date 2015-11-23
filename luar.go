package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

// New creates and returns a new lua.LValue for the given value.
//
// The following types are supported:
//  Kind            gopher-lua Type
//  -------------------------------
//  nil             LNil
//  Bool            LBool
//  Int             LNumber
//  Int8            LNumber
//  Int16           LNumber
//  Int32           LNumber
//  Int64           LNumber
//  Uint            LNumber
//  Uint8           LNumber
//  Uint32          LNumber
//  Uint64          LNumber
//  Float32         LNumber
//  Float64         LNumber
//  Complex64       *LUserData
//  Complex128      *LUserData
//  Array           *LUserData
//  Chan            *LUserData
//  Interface       *LUserData
//  Func            *lua.LFunction
//  Map             *LUserData
//  Ptr             *LUserData
//  Slice           *LUserData
//  String          LString
//  Struct          *LUserData
//  UnsafePointer   *LUserData
func New(L *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}
	if lval, ok := value.(lua.LValue); ok {
		return lval
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if val.IsNil() {
			return lua.LNil
		}
	}

	switch val.Kind() {
	case reflect.Bool:
		return lua.LBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(float64(val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(float64(val.Uint()))
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(val.Float())
	case reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Struct:
		ud := L.NewUserData()
		ud.Value = val.Interface()
		ud.Metatable = getMetatable(L, val)
		return ud
	case reflect.Func:
		return funcWrapper(L, val)
	case reflect.Interface:
		ud := L.NewUserData()
		ud.Value = val.Interface()
		return ud
	case reflect.String:
		return lua.LString(val.String())
	default:
		ud := L.NewUserData()
		ud.Value = val.Interface()
		return ud
	}
}

// NewType returns a new type creator for the given value's type.
//
// When the lua.LValue is called, a new value will be created that is the
// same type as value's type.
func NewType(L *lua.LState, value interface{}) lua.LValue {
	val := reflect.TypeOf(value)
	ud := L.NewUserData()
	ud.Value = val
	ud.Metatable = getTypeMetatable(L, val)

	return ud
}

func lValueToReflect(v lua.LValue, hint reflect.Type) reflect.Value {
	if hint == refTypeLuaLValue {
		return reflect.ValueOf(v)
	}
	switch converted := v.(type) {
	case lua.LBool:
		return reflect.ValueOf(bool(converted))
	case lua.LChannel:
		return reflect.ValueOf(converted)
	case lua.LNumber:
		return reflect.ValueOf(converted).Convert(hint)
	case *lua.LFunction:
		return reflect.ValueOf(converted)
	case *lua.LNilType:
		return reflect.Zero(hint)
	case *lua.LState:
		return reflect.ValueOf(converted)
	case lua.LString:
		return reflect.ValueOf(string(converted))
	case *lua.LTable:
		return reflect.ValueOf(converted)
	case *lua.LUserData:
		return reflect.ValueOf(converted.Value)
	}
	panic("fatal lValueToReflect error")
}
