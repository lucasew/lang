package language

// TagalogRelevantRuleIDs ports Tagalog.getRelevantRules rule IDs (class getId only).
func TagalogRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"MORFOLOGIK_RULE_TL",
	}
}
