package uk

import "strings"

// PosTagHelper ports tagging.uk.PosTagHelper helpers for Ukrainian POS tags.
type PosTagHelper struct{}

// NoVidminokSubstr ports PosTagHelper.NO_VIDMINOK_SUBSTR.
const NoVidminokSubstr = ":nv"

// VidminkyMap ports PosTagHelper.VIDMINKY_MAP (case code → Ukrainian name).
// Iteration order for messages matches Java LinkedHashMap insertion.
var VidminkyMap = map[string]string{
	"v_naz": "називний",
	"v_rod": "родовий",
	"v_dav": "давальний",
	"v_zna": "знахідний",
	"v_oru": "орудний",
	"v_mis": "місцевий",
	"v_kly": "кличний",
}

// VidminkyIMap ports PosTagHelper.VIDMINKY_I_MAP (includes v_inf for verb gov messages).
var VidminkyIMap = map[string]string{
	"v_naz": "називний",
	"v_rod": "родовий",
	"v_dav": "давальний",
	"v_zna": "знахідний",
	"v_oru": "орудний",
	"v_mis": "місцевий",
	"v_kly": "кличний",
	"v_inf": "інфінітив",
}

// VidminkyOrder is LinkedHashMap insertion order for VIDMINKY_MAP.
var VidminkyOrder = []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis", "v_kly"}

// VidminokName returns the Ukrainian case name for a v_* code, or the code itself.
func VidminokName(code string) string {
	if n, ok := VidminkyMap[code]; ok {
		return n
	}
	return code
}

// VidminokIName returns VIDMINKY_I_MAP name (incl. інфінітив).
func VidminokIName(code string) string {
	if n, ok := VidminkyIMap[code]; ok {
		return n
	}
	return code
}

// GenderMap ports PosTagHelper.GENDER_MAP.
var GenderMap = map[string]string{
	"m": "ч.р.",
	"f": "ж.р.",
	"n": "с.р.",
	"p": "мн.",
	"s": "одн.",
	"i": "інф.",
	"o": "безос. форма",
}

// GenderName returns the Ukrainian gender label, or the code itself.
func GenderName(code string) string {
	if n, ok := GenderMap[code]; ok {
		return n
	}
	return code
}

// PersonMap ports PosTagHelper.PERSON_MAP.
var PersonMap = map[string]string{
	"1": "1-а особа",
	"2": "2-а особа",
	"3": "3-я особа",
}

// PersonName returns PERSON_MAP label or the code itself.
func PersonName(code string) string {
	if n, ok := PersonMap[code]; ok {
		return n
	}
	return code
}

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
