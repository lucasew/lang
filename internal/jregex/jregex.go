// Package jregex adapts Java Pattern syntax fragments to Go RE2.
package jregex

import (
	"regexp"
	"strconv"
	"strings"
)

// Compile compiles a Java-ish regex anchored as full match when fullMatch is true.
func Compile(pat string, caseSensitive bool) (*regexp.Regexp, error) {
	pat = JavaToGo(pat)
	flags := "(?m"
	if !caseSensitive {
		flags += "i"
	}
	flags += ")"
	return regexp.Compile(flags + "^(?:" + pat + ")$")
}

// CompileSearch compiles without full-string anchors.
func CompileSearch(pat string) (*regexp.Regexp, error) {
	return regexp.Compile("(?m:" + JavaToGo(pat) + ")")
}

// JavaToGo converts \uXXXX escapes (and a few common differences).
func JavaToGo(pat string) string {
	var b strings.Builder
	b.Grow(len(pat) + 8)
	for i := 0; i < len(pat); i++ {
		if pat[i] == '\\' && i+1 < len(pat) {
			if pat[i+1] == 'u' && i+5 < len(pat) {
				hex := pat[i+2 : i+6]
				if _, err := strconv.ParseUint(hex, 16, 32); err == nil {
					b.WriteString(`\x{`)
					b.WriteString(hex)
					b.WriteByte('}')
					i += 5
					continue
				}
			}
		}
		b.WriteByte(pat[i])
	}
	return b.String()
}
