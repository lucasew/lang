package de

// VerbPrefixes ports org.languagetool.tagging.de.VerbPrefixes.
var verbPrefixes = []string{
	"ab", "an", "auf", "aus", "auseinander", "bei", "ein", "empor", "entgegen", "entlang", "entzwei",
	"fehl", "fern", "fest", "fort", "gegenüber", "heim", "hinterher", "hoch", "los", "mit", "nach", "neben", "nieder", "vor",
	"weg", "weiter", "zu", "zurecht", "zurück", "zusammen", "da", "hin", "her",
	"herab", "heran", "herauf", "heraus", "herbei", "herein", "hernieder", "herüber", "herum", "herunter", "hervor", "herzu",
	"hinab", "hinan", "hinauf", "hinaus", "hinein", "hinüber", "hinunter", "hinweg", "hinzu", "vorab", "voran", "vorauf", "voraus",
	"vorbei", "vorweg", "vorher", "vorüber",
	"dabei", "dafür", "dagegen", "daher", "dahin", "dahinter", "daneben", "daran", "darauf", "darein", "darüber", "darunter",
	"hinter", "dran", "drauf", "drein", "drüber", "drunter",
	"davon", "davor", "dazu", "dazwischen",
}

// GetVerbPrefixes returns a copy of the known separable/inseparable prefix list.
func GetVerbPrefixes() []string {
	return append([]string(nil), verbPrefixes...)
}

// IsVerbPrefix reports whether p is a known verb prefix (case-sensitive lowercase).
func IsVerbPrefix(p string) bool {
	for _, x := range verbPrefixes {
		if x == p {
			return true
		}
	}
	return false
}

// VerbPrefixes is the Java-name twin for the prefix list helpers.
type VerbPrefixes struct{}

// Get returns a copy of known verb prefixes.
func (VerbPrefixes) Get() []string { return GetVerbPrefixes() }

// Contains reports whether p is a known prefix.
func (VerbPrefixes) Contains(p string) bool { return IsVerbPrefix(p) }
