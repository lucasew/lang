package language

// Belarusian rule priorities from org.languagetool.language.Belarusian.getPriorityForId.
// Switch-table port as map (exact id equality). Java is king — do not invent IDs.

var belarusianPriorityExact = map[string]int{
	"RUSSIAN_SIMPLE_REPLACE_RULE": 10,
	"BELARUSIAN_SPECIFIC_CASE":    9,
	"Word_root_repeat":            -1,
	"PUNCT_DPT_2":                 -2,
	"TOO_LONG_PARAGRAPH":          -15,
}

// BelarusianPriorityExactMap returns a defensive copy of the exact-id table.
func BelarusianPriorityExactMap() map[string]int {
	out := make(map[string]int, len(belarusianPriorityExact))
	for k, v := range belarusianPriorityExact {
		out[k] = v
	}
	return out
}

// BelarusianPriorityForId ports Belarusian.getPriorityForId (then Language base).
func BelarusianPriorityForId(id string) int {
	if p, ok := belarusianPriorityExact[id]; ok {
		return p
	}
	return languagePriorityForId(id)
}
