package uk

// lenPrefixBytes maps a lowercased prefix length onto the original surface in bytes.
func lenPrefixBytes(surface, prefixLower string) int {
	pr := []rune(prefixLower)
	sr := []rune(surface)
	if len(pr) > len(sr) {
		return 0
	}
	return len(string(sr[:len(pr)]))
}
