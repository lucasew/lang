package language

// LithuanianRelevantRuleIDs ports Lithuanian.getRelevantRules rule IDs (class getId only).
func LithuanianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_LT_LT",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
	}
}
