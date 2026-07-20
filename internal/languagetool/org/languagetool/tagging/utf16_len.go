package tagging

import "unicode/utf16"

// UTF16Len ports Java String.length() (UTF-16 code units).
// BaseTagger and language taggers advance startPos with word.length().
func UTF16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}
