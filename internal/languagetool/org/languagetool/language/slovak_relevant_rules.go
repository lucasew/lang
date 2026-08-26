package language

// SlovakRelevantRuleIDs ports Slovak.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Slovak.java (SlovakVesRule commented out — omit).
func SlovakRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (SK symbols)
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		// specific to Slovak:
		"SK_COMPOUNDS",          // CompoundRule
		"MORFOLOGIK_RULE_SK_SK", // MorfologikSlovakSpellerRule
		// SlovakVesRule commented out in Java — omit
	}
}

// GetRelevantRuleIDs ports Slovak.getRelevantRules on SlovakLang.
func (v SlovakLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), SlovakRelevantRuleIDs()...)
}
