package luar

import (
	"github.com/yuin/gopher-lua"
)

// MetaCall is a struct or struct pointer that defines a fallback action for
// the Lua __call metamethod.
type MetaCall interface {
	LuarCall(...lua.LValue) ([]lua.LValue, error)
}

// MetaIndex is a struct or struct pointer that defines a fallback action for
// the Lua __index metamethod.
type MetaIndex interface {
	LuarIndex(key lua.LValue) (lua.LValue, error)
}

// MetaNewIndex is a struct or struct pointer that defines a fallback action
// for the Lua __newindex metamethod.
type MetaNewIndex interface {
	LuarNewIndex(key, value lua.LValue) error
}
