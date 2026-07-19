package language

// CatalanRelevantRuleIDs ports Catalan.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Catalan.java.
// ValencianCatalan adds WordCoherencyValencianRule — see GetRelevantRuleIDs.
func CatalanRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		// CatalanUnpairedBracketsRule: getId() override commented out in Java →
		// GenericUnpairedBracketsRule default id UNPAIRED_BRACKETS (not CA_UNPAIRED_BRACKETS).
		"UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"TOO_LONG_SENTENCE",
		"CATALAN_WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_CA_ES",
		"CA_UNPAIRED_QUESTION",
		"CA_UNPAIRED_EXCLAMATION",
		"CATALAN_WRONG_WORD_IN_CONTEXT",
		"CA_SIMPLE_REPLACE_VERBS",
		"CA_SIMPLE_REPLACE_BALEARIC",
		"CA_SIMPLE_REPLACE_SIMPLE",
		"CA_SIMPLE_REPLACE_MULTIWORDS",
		"NOMS_OPERACIONS",
		"CA_SIMPLE_REPLACE_DIACRITICS_IEC",
		"CA_SIMPLE_REPLACE_ANGLICISM",
		"PRONOMS_FEBLES_DUPLICATS",
		"CA_CHECKCASE",
		"ADVERBIS_MENT",
		"CATALAN_WORD_REPEAT_BEGINNING_RULE",
		"CA_COMPOUNDS",
		// CatalanRepeatedWordsRule is not in Java getRelevantRules asList (omit).
		"CA_SIMPLE_REPLACE_DNV",
		"CA_SIMPLE_REPLACE_DNV_COLLOQUIAL",
		"CA_SIMPLE_REPLACE_DNV_SECONDARY",
		"CA_WORD_COHERENCY",
		"PUNCTUATION_PARAGRAPH_END",
		"CA_REMOTE_RULE",
		"CA_SPLIT_LONG_SENTENCE",
		"IGNORE_PROPER_NOUNS",
	}
}

// GetRelevantRuleIDs ports Catalan.getRelevantRules (+ Valencian coherency rule).
func (v CatalanVariant) GetRelevantRuleIDs() []string {
	out := append([]string(nil), CatalanRelevantRuleIDs()...)
	if v.Valencian {
		// Java ValencianCatalan.getRelevantRules: super + WordCoherencyValencianRule
		out = append(out, "CA_WORD_COHERENCY_VALENCIA")
	}
	return out
}
