package luar

import (
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
