package luar

import (
	"unicode"
	"unicode/utf8"
)

var (
	// If set to true (in an init function or something), then globally,
	// all reflected structs will allow access to unexported members through
	// luar. You really shouldn't set this to true unless you know for sure
	// all usages of luar will be for debugging purposes.
	// TODO: This is so gross. We should move this setting to a registered
	// object specific place, instead of being global and terrible. Should
	// probably be a different type of New method or something.
	AllowUnexportedAccesses bool = false
)

func getExportedName(name string) string {
	if AllowUnexportedAccesses {
		return name
	}
	buf := []byte(name)
	first, n := utf8.DecodeRune(buf)
	if n == 0 {
		return name
	}
	return string(unicode.ToUpper(first)) + string(buf[n:])
}
