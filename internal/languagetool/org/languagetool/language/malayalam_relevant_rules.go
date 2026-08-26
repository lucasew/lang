package language

// MalayalamRelevantRuleIDs ports Malayalam.getRelevantRules rule IDs (class getId only).
func MalayalamRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_ML_IN",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
	}
}
