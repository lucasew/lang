package language

// SimpleGermanRelevantRuleIDs ports SimpleGerman.getRelevantRules.
// Java only registers de.LongSentenceRule (not full GermanyGerman list).
func SimpleGermanRelevantRuleIDs() []string {
	return []string{
		"TOO_LONG_SENTENCE_DE",
	}
}

// SimpleGermanShortCode ports SimpleGerman.getShortCode (BCP47 private-use tag).
const SimpleGermanShortCode = "de-DE-x-simple-language"

// SimpleGermanGetRuleFileNames ports SimpleGerman.getRuleFileNames.
// Java does not call super: only shortCode/grammar.xml for this private-use code.
func SimpleGermanGetRuleFileNames() []string {
	return []string{
		"/org/languagetool/rules/" + SimpleGermanShortCode + "/grammar.xml",
	}
}

// SimpleGermanIsVariant ports SimpleGerman.isVariant() → true.
func SimpleGermanIsVariant() bool { return true }

// SimpleGermanGetName ports SimpleGerman.getName.
func SimpleGermanGetName() string { return "Simple German" }

// SimpleGermanGetMaintainers ports SimpleGerman.getMaintainers → Annika Nietzio.
func SimpleGermanGetMaintainers() []Contributor {
	return []Contributor{NewContributor("Annika Nietzio")}
}
