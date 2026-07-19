package language

// DutchRelevantRuleIDs ports Dutch.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Dutch.java.
func DutchRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS", // GenericUnpairedBracketsRule default id
		"UPPERCASE_SENTENCE_START",
		"MORFOLOGIK_RULE_NL_NL",
		"WHITESPACE_RULE",
		"NL_COMPOUNDS",
		"DUTCH_WRONG_WORD_IN_CONTEXT",
		"NL_WORD_COHERENCY",
		"NL_SIMPLE_REPLACE",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"NL_PREFERRED_WORD_RULE",
		"NL_SPACE_IN_COMPOUND",
		"SENTENCE_WHITESPACE",
		"NL_CHECKCASE",
	}
}

// GetRelevantRuleIDs ports Dutch.getRelevantRules for Dutch variants.
func (v DutchVariant) GetRelevantRuleIDs() []string {
	return append([]string(nil), DutchRelevantRuleIDs()...)
}
