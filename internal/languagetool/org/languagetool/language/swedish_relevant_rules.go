package language

// SwedishRelevantRuleIDs ports Swedish.getRelevantRules rule IDs (class getId only).
func SwedishRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"HUNSPELL_RULE",
		"TOO_LONG_PARAGRAPH",
		"UPPERCASE_SENTENCE_START",
		"TOO_LONG_SENTENCE",
		"WORD_REPEAT_RULE",
		"SV_WORD_COHERENCY",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"SV_COMPOUNDS",
	}
}
