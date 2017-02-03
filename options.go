package luar // import "layeh.com/gopher-luar"

// NameOptions is a type to define the valid settings for how to handle names.
type NameOptions int8

// Options defines the configurable value for the luar library.
type Options struct {
	MethodNames, FieldNames NameOptions
}

// Specifies the method to use in creating names for fields and methods for
// Go types in Lua.
const (
	ExportedUnexportNameStyle NameOptions = iota
	ExportedNameStyle
	UnexportedNameStyle
)

// DefaultOptions for the library
var DefaultOptions = Options{
	MethodNames: ExportedUnexportNameStyle,
	FieldNames:  ExportedUnexportNameStyle,
}

// stores current configuration
var options = DefaultOptions

// Configure lets the user pass in a customer configuration to adjust the
// behavior of the library.
func Configure(opts Options) {
	options = opts
}
