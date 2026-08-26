package implementation

import "unicode/utf16"

// javaStringLen ports Java String.length() (UTF-16 code units).
func javaStringLen(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// javaChars ports Java String as UTF-16 code units.
func javaChars(s string) []uint16 {
	return utf16.Encode([]rune(s))
}

// javaFromChars rebuilds a Go string from UTF-16 units.
func javaFromChars(u []uint16) string {
	if len(u) == 0 {
		return ""
	}
	return string(utf16.Decode(u))
}

// javaSubstring ports String.substring(from, to) with UTF-16 indices.
func javaSubstring(s string, from, to int) string {
	u := javaChars(s)
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return javaFromChars(u[from:to])
}

// javaDeleteCharAt ports StringBuilder.deleteCharAt(i) on UTF-16 units.
func javaDeleteCharAt(s string, i int) string {
	u := javaChars(s)
	if i < 0 || i >= len(u) {
		return s
	}
	out := make([]uint16, 0, len(u)-1)
	out = append(out, u[:i]...)
	out = append(out, u[i+1:]...)
	return javaFromChars(out)
}

// javaIndexOfChar ports String.indexOf(char) for a UTF-16 code unit.
func javaIndexOfChar(s string, ch uint16) int {
	u := javaChars(s)
	for i, c := range u {
		if c == ch {
			return i
		}
	}
	return -1
}

// javaCharAt ports String.charAt(i) for a UTF-16 code unit index (panics if OOB like Java).
func javaCharAt(s string, i int) uint16 {
	u := javaChars(s)
	return u[i]
}

// javaStringHashCode ports Java String.hashCode():
// s[0]*31^(n-1) + s[1]*31^(n-2) + ... over UTF-16 code units, 32-bit int overflow.
func javaStringHashCode(s string) int {
	var h int32
	for _, cu := range javaChars(s) {
		h = 31*h + int32(cu)
	}
	return int(h)
}
