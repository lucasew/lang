package language

// SpanishRelevantRuleIDs ports Spanish.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Spanish.java.
func SpanishRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"ES_UNPAIRED_BRACKETS",
		"ES_QUESTION_MARK",
		"MORFOLOGIK_RULE_ES",
		"UPPERCASE_SENTENCE_START",
		"SPANISH_WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
		"ES_WIKIPEDIA_COMMON_ERRORS",
		"SPANISH_WRONG_WORD_IN_CONTEXT",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"ES_SIMPLE_REPLACE_SIMPLE",
		"ES_SIMPLE_REPLACE_VERBS",
		"SPANISH_WORD_REPEAT_BEGINNING_RULE",
		"ES_COMPOUNDS",
		"ES_REPEATEDWORDS",
	}
}

// GetRelevantRuleIDs ports Spanish.getRelevantRules for Spanish variants.
func (v SpanishVariant) GetRelevantRuleIDs() []string {
	return append([]string(nil), SpanishRelevantRuleIDs()...)
}
