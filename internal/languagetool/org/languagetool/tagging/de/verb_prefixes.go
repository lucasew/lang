package de

// VerbPrefixes ports org.languagetool.tagging.de.VerbPrefixes.
// Order: alphabetical then longest-first (Java static initializer).
var verbPrefixes = []string{
	"auseinander", "aneinander", "dazwischen", "gegenüber", "hernieder", "hinterher", "dahinter", "darunter",
	"entgegen", "herunter", "hinunter", "zusammen", "dagegen", "daneben", "darüber", "drunter",
	"entlang", "entzwei", "herüber", "hinüber", "vorüber", "zurecht", "darauf", "darein",
	"drüber", "herauf", "heraus", "herbei", "herein", "hervor", "hinauf", "hinaus",
	"hinein", "hinter", "hinweg", "nieder", "runter", "vorauf", "voraus", "vorbei",
	"vorher", "vorweg", "weiter", "wieder", "zurück", "dabei", "dafür", "daher",
	"dahin", "daran", "davon", "davor", "drauf", "drein", "durch", "empor",
	"gegen", "herab", "heran", "herum", "herzu", "hinab", "hinan", "hinzu",
	"neben", "rüber", "umher", "unter", "vorab", "voran", "wider", "dazu",
	"dran", "fehl", "fern", "fest", "fort", "frei", "heim", "hoch",
	"nach", "rauf", "rein", "über", "auf", "aus", "bei", "ein",
	"ent", "her", "hin", "los", "mit", "ran", "ver", "vor",
	"weg", "zer", "ab", "an", "da", "um", "zu",
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
