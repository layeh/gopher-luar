package luar

import (
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

// Config is used to define luar behaviour for a particular *lua.LState.
type Config struct {
	// The name generating function that defines under which names Go
	// struct fields will be accessed.
	//
	// If nil, the default behaviour is used:
	//   - if the "luar" tag of the field is "", the field name and its name
	//     with a lowercase first letter is returned
	//  - if the tag is "-", no name is returned (i.e. the field is not
	//    accessible)
	//  - for any other tag value, that value is returned
	FieldNames func(s reflect.Type, f reflect.StructField) []string

	// The name generating function that defines under which names Go
	// methods will be accessed.
	//
	// If nil, the default behaviour is used:
	//   - the method name and its name with a lowercase first letter
	MethodNames func(t reflect.Type, m reflect.Method) []string

	// The metatable post-processor function. This gets run last, after the
	// [default implementation] is done. If nil, skipped. You may use this to
	// provide custom metamethods or the like for specific Go types.
	//
	// If the constructor parameter is true, the metatable is being created for
	// the global constructor metatable; as such, this will only be called once.
	//
	// [default implementation]: https://github.com/layeh/gopher-luar/blob/master/cache.go#L92
	Metatable func(L *lua.LState, t reflect.Type, mt *lua.LTable, constructor bool) *lua.LTable

	// When true, all metatables are fully processed as soon as they are
	// discovered. This increases memory use but enables operations that depend
	// on complete metatable information (e.g., full annotation generation via
	// the use of [Config.Metatable]).
	//
	// When false (default), metatables are processed only when needed. Calling
	// [New] triggers initial processing, and further processing occurs only
	// when a field, method, or call is accessed. If true, all referenced types
	// are processed during New rather than on first use.
	//
	// This is not affected by [NewType] due to how luar is implemented. If you
	// wish to do things based on the type of a [NewType], simply call [New] and
	// unless you need it, discard the result.
	PreprocessMetatables bool

	// overhead: to prevent recursion for [Config.PreprocessMetatables], we must
	// store a map of types that are currently being processed. I wanted to use
	// [Config.regular] with an empty table, but [getMetatable] puts the exact
	// number of entries as the capacity, and unless I make a second switch,
	// I can't think of a good way to emulate this behaviour. This will work for
	// now, I think. FIXME.
	processing map[reflect.Type]bool
	regular    map[reflect.Type]*lua.LTable
	types      *lua.LTable
}

func newConfig() *Config {
	return &Config{
		processing: make(map[reflect.Type]bool),
		regular:    make(map[reflect.Type]*lua.LTable),
	}
}

// GetConfig returns the luar configuration options for the given *lua.LState.
func GetConfig(L *lua.LState) *Config {
	const registryKey = "github.com/layeh/gopher-luar"

	registry := L.Get(lua.RegistryIndex).(*lua.LTable)
	lConfig, ok := registry.RawGetString(registryKey).(*lua.LUserData)
	if !ok {
		lConfig = L.NewUserData()
		lConfig.Value = newConfig()
		registry.RawSetString(registryKey, lConfig)
	}
	return lConfig.Value.(*Config)
}

func defaultFieldNames(s reflect.Type, f reflect.StructField) []string {
	const tagName = "luar"

	tag := f.Tag.Get(tagName)
	if tag == "-" {
		return nil
	}
	if tag != "" {
		return []string{tag}
	}
	return []string{
		f.Name,
		getUnexportedName(f.Name),
	}
}

func defaultMethodNames(t reflect.Type, m reflect.Method) []string {
	return []string{
		m.Name,
		getUnexportedName(m.Name),
	}
}
