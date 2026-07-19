package language

// TamilRelevantRuleIDs ports Tamil.getRelevantRules rule IDs (class getId only).
func TamilRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"WHITESPACE_RULE",
		"TOO_LONG_SENTENCE",
		"SENTENCE_WHITESPACE",
	}
}
