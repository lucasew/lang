package language

// ItalianRelevantRuleIDs ports Italian.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Italian.java.
func ItalianRelevantRuleIDs() []string {
	return []string{
		"WHITESPACE_PUNCTUATION",
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"MORFOLOGIK_RULE_IT_IT",
		"UPPERCASE_SENTENCE_START",
		"ITALIAN_WORD_REPEAT_RULE",
		"WHITESPACE_RULE",
	}
}

// ItalianGetRelevantRuleIDs is the package helper (Italian is ItalianLang value).
func ItalianGetRelevantRuleIDs() []string {
	return append([]string(nil), ItalianRelevantRuleIDs()...)
}

// GetRelevantRuleIDs ports Italian.getRelevantRules on ItalianLang.
func (v ItalianLang) GetRelevantRuleIDs() []string {
	return ItalianGetRelevantRuleIDs()
}
