package language

// SimpleGermanPriorityForId ports org.languagetool.language.SimpleGerman.getPriorityForId.
// SimpleGerman extends GermanyGerman → super is GermanPriorityForId.
// Java is king — do not invent IDs.
func SimpleGermanPriorityForId(id string) int {
	switch id {
	case "TOO_LONG_SENTENCE":
		return 10
	case "LANGES_WORT":
		return -1
	}
	return GermanPriorityForId(id)
}

// isSimpleGermanCode reports de-DE-x-simple-language private-use tags.
func isSimpleGermanCode(langCode string) bool {
	switch langCode {
	case "de-DE-x-simple-language", "de-de-x-simple-language":
		return true
	}
	return false
}

// GermanPriorityForIdForCode selects SimpleGerman or German by language code.
func GermanPriorityForIdForCode(langCode string) func(string) int {
	if isSimpleGermanCode(langCode) {
		return SimpleGermanPriorityForId
	}
	return GermanPriorityForId
}
