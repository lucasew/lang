package uk

import "strings"

// PosTagHelper ports tagging.uk.PosTagHelper helpers for Ukrainian POS tags.
type PosTagHelper struct{}

// HasPos reports whether pos contains the given tag fragment (colon-separated).
func HasPos(pos, fragment string) bool {
	if pos == "" || fragment == "" {
		return false
	}
	for _, p := range strings.Split(pos, ":") {
		if p == fragment {
			return true
		}
	}
	return false
}

// IsNoun reports noun tags (noun:...).
func IsNoun(pos string) bool {
	return strings.HasPrefix(pos, "noun")
}

// IsVerb reports verb tags.
func IsVerb(pos string) bool {
	return strings.HasPrefix(pos, "verb")
}

// IsAdj reports adjective tags.
func IsAdj(pos string) bool {
	return strings.HasPrefix(pos, "adj")
}

// Gender returns m/f/n/s from POS if present.
func Gender(pos string) string {
	for _, g := range []string{"m", "f", "n", "s", "p"} {
		if HasPos(pos, g) {
			return g
		}
	}
	return ""
}

// Case returns nom/gen/dat/acc/ins/loc/voc if present.
func Case(pos string) string {
	for _, c := range []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly", "nom", "gen", "dat", "acc", "ins", "loc", "voc"} {
		if HasPos(pos, c) {
			return c
		}
	}
	return ""
}
