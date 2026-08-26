package language

// IcelandicRelevantRuleIDs ports Icelandic.getRelevantRules rule IDs (class getId only).
func IcelandicRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"HUNSPELL_NO_SUGGEST_RULE",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
	}
}
