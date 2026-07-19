package language

// CrimeanTatarRelevantRuleIDs ports CrimeanTatar.getRelevantRules rule IDs (class getId only).
func CrimeanTatarRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"MORFOLOGIK_RULE_CRH_UA",
	}
}
