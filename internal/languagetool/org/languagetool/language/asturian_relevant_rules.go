package language

// AsturianRelevantRuleIDs ports Asturian.getRelevantRules rule IDs (class getId only).
func AsturianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_AST",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
	}
}
