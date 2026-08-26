package language

// PolishRelevantRuleIDs ports Polish.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Polish.java.
func PolishRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"UPPERCASE_SENTENCE_START",
		"WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"PL_UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_PL_PL",
		"PL_WORD_REPEAT",
		"PL_COMPOUNDS",
		"PL_SIMPLE_REPLACE",
		"PL_WORD_COHERENCY",
		"DASH_RULE", // AbstractDashRule / pl.DashRule
	}
}

// GetRelevantRuleIDs ports Polish.getRelevantRules on PolishLang.
func (v PolishLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), PolishRelevantRuleIDs()...)
}
