package language

// Russian rule priorities from org.languagetool.language.Russian.getPriorityForId.
// Switch-table port as map (exact id equality). Java is king — do not invent IDs.

var russianPriorityExact = map[string]int{
	"RU_DASH_RULE":                12,
	"RU_COMPOUNDS":                11,
	"RUSSIAN_SIMPLE_REPLACE_RULE": 10,
	"RUSSIAN_SPECIFIC_CASE":       9,
	"MORFOLOGIC_RULE_RU_RU_YO":    2,
	"MORFOLOGIC_RULE_RU_RU":       1,
	"Word_root_repeat":            -1,
	"PUNCT_DPT_2":                 -2,
	"TOO_LONG_PARAGRAPH":          -15,
}

// RussianPriorityExactMap returns a defensive copy of the exact-id table.
func RussianPriorityExactMap() map[string]int {
	out := make(map[string]int, len(russianPriorityExact))
	for k, v := range russianPriorityExact {
		out[k] = v
	}
	return out
}

// RussianPriorityForId ports Russian.getPriorityForId (then Language base).
func RussianPriorityForId(id string) int {
	if p, ok := russianPriorityExact[id]; ok {
		return p
	}
	return languagePriorityForId(id)
}
