package language

// FrenchRelevantRuleIDs ports French.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in French.java.
func FrenchRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"FR_SPELLING_RULE",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"FR_COMPOUNDS",
		"FRENCH_WHITESPACE_STRICT",
		"FRENCH_WHITESPACE",
		"FR_SIMPLE_REPLACE_SIMPLE",
		"FR_REPEATEDWORDS",
	}
}

// GetRelevantRuleIDs ports French.getRelevantRules for French variants.
func (v FrenchVariant) GetRelevantRuleIDs() []string {
	return append([]string(nil), FrenchRelevantRuleIDs()...)
}
