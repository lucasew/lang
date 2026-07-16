package de

// VerbPrefixes ports tagging.de.VerbPrefixes (separable verb prefixes subset).
var VerbPrefixes = []string{
	"ab", "an", "auf", "aus", "bei", "dar", "ein", "empor", "entgegen",
	"entlang", "entzwei", "fehl", "fern", "fest", "fort", "frei", "gegenüber",
	"gleich", "heim", "her", "herab", "heran", "herauf", "heraus", "herbei",
	"herein", "herum", "herunter", "hervor", "hin", "hinab", "hinan", "hinauf",
	"hinaus", "hinein", "hinten", "hinterher", "hinunter", "hinweg", "hinzu",
	"hoch", "los", "mit", "nach", "neben", "nieder", "statt", "um", "vor",
	"voran", "voraus", "vorbei", "vorher", "vorüber", "weg", "weiter", "wieder",
	"zu", "zurecht", "zurück", "zusammen",
}

// IsVerbPrefix reports whether s is a known separable prefix.
func IsVerbPrefix(s string) bool {
	for _, p := range VerbPrefixes {
		if p == s {
			return true
		}
	}
	return false
}
