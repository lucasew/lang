package language

// IrishPriorityForId ports org.languagetool.language.Irish.getPriorityForId.
// Java is king — do not invent IDs.
func IrishPriorityForId(id string) int {
	if id == "TOO_LONG_PARAGRAPH" {
		return -15
	}
	return languagePriorityForId(id)
}
