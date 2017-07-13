// Package luar provides custom type reflection to gopher-lua
// (https://github.com/yuin/gopher-lua).
//
// Notice
//
// This package is currently in development, and its behavior may change. This
// message will be removed once the package is considered stable.
//
// Lua to Go conversions
//
// The Lua types are automatically converted to match the output Go type, as
// described below:
//
//  Lua type    Go kind/type
//  -----------------------------------------------------
//  LBool       bool
//              string ("true" or "false")
//  LChannel    chan lua.LValue
//  LNumber     numeric value
//              string (strconv.Itoa)
//  LFunction   func
//  LNilType    chan, func, interface, map, ptr, slice, unsafe pointer
//  LState      *lua.LState
//  LString     string
//  LTable      slice
//              map
//              struct
//              *struct
//  LUserData   underlying lua.LUserData.Value type
//
// Example creating a Go slice from Lua:
//  type Group struct {
//      Names []string
//  }
//
//  g := new(Group)
//  L.SetGlobal("g", luar.New(L, g))
//  ---
//  g.Names = {"Tim", "Frank", "George"}
//
// Thread safety
//
// This package accesses and modifies the Lua state's registry. This happens
// when functions like New are called, and potentially when luar-created values
// are used. It is your responsibility to ensure that concurrent access of the
// state's registry does not happen.
package luar // import "layeh.com/gopher-luar"
