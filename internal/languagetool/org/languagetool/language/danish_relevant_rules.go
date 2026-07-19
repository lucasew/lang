package language

// DanishRelevantRuleIDs ports Danish.getRelevantRules rule IDs (class getId only).
func DanishRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"HUNSPELL_RULE",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
	}
}
