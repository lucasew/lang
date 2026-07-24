package language

// BretonRelevantRuleIDs ports Breton.getRelevantRules rule IDs (class getId only).
func BretonRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"MORFOLOGIK_RULE_BR_FR",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"BR_TOPO",
	}
}
