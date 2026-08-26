package language

// Galician rule priorities from org.languagetool.language.Galician.getPriorityForId.
// Switch-table port as map (exact id equality). Commented Java cases omitted — not invent.

var galicianPriorityExact = map[string]int{
	"DEGREE_MINUTES_SECONDS":     30,
	"UNPAIRED_BRACKETS":          -5,
	"GL_BARBARISM_REPLACE":       -10,
	"GL_SIMPLE_REPLACE":          -11,
	"GL_REDUNDANCY_REPLACE":      -12,
	"GL_WORDINESS_REPLACE":       -13,
	"TOO_LONG_PARAGRAPH":         -15,
	"GL_WIKIPEDIA_COMMON_ERRORS": -45,
	"HUNSPELL_RULE":              -50,
	"REPEATED_WORDS":             -210,
	"REPEATED_WORDS_3X":          -211,
	"TOO_LONG_SENTENCE_20":       -997,
	"TOO_LONG_SENTENCE_25":       -998,
	"TOO_LONG_SENTENCE_30":       -999,
	"TOO_LONG_SENTENCE_35":       -1000,
	"TOO_LONG_SENTENCE_40":       -1001,
	"TOO_LONG_SENTENCE_45":       -1002,
	"TOO_LONG_SENTENCE_50":       -1003,
	"TOO_LONG_SENTENCE_60":       -1004,
}

// GalicianPriorityExactMap returns a defensive copy of the exact-id table.
func GalicianPriorityExactMap() map[string]int {
	out := make(map[string]int, len(galicianPriorityExact))
	for k, v := range galicianPriorityExact {
		out[k] = v
	}
	return out
}

// GalicianPriorityForId ports Galician.getPriorityForId (then Language base).
func GalicianPriorityForId(id string) int {
	if p, ok := galicianPriorityExact[id]; ok {
		return p
	}
	return languagePriorityForId(id)
}
