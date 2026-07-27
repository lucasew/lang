package language

// TagalogRelevantRuleIDs ports Tagalog.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Tagalog.java exactly.
func TagalogRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (default symbols)
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"MORFOLOGIK_RULE_TL",           // MorfologikTagalogSpellerRule
	}
}

// GetRelevantRuleIDs ports Tagalog.getRelevantRules on TagalogLang.
func (v TagalogLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), TagalogRelevantRuleIDs()...)
}
