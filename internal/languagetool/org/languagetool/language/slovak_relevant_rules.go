package language

// SlovakRelevantRuleIDs ports Slovak.getRelevantRules rule IDs (class getId only).
func SlovakRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
		"SK_COMPOUNDS",
		"MORFOLOGIK_RULE_SK_SK",
	}
}
