package language

// SlovenianRelevantRuleIDs ports Slovenian.getRelevantRules rule IDs (class getId only).
func SlovenianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_SL_SI",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
	}
}
