package luar

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

func getExportedName(name string) string {
	buf := []byte(name)
	first, n := utf8.DecodeRune(buf)
	if n == 0 {
		return name
	}
	return string(unicode.ToUpper(first)) + string(buf[n:])
}

func getString(i interface{}) (string, error) {
	if stringer, ok := i.(fmt.Stringer); ok {
		return stringer.String(), nil
	}
	return reflect.ValueOf(i).String(), nil
}
