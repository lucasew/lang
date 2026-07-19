package language

// EsperantoRelevantRuleIDs ports Esperanto.getRelevantRules rule IDs (class getId only).
func EsperantoRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"HUNSPELL_RULE",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
	}
}
