package language

// PolishPriorityForId ports org.languagetool.language.Polish.getPriorityForId.
// Java is king — do not invent IDs.
func PolishPriorityForId(id string) int {
	// so that it does not override more important rules
	if id == "ZDANIA_ZLOZONE" {
		return -1
	}
	return languagePriorityForId(id)
}
