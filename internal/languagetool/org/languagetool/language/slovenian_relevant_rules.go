package language

// SlovenianRelevantRuleIDs ports Slovenian.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Slovenian.java exactly.
func SlovenianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (SL symbols)
		"MORFOLOGIK_RULE_SL_SI",        // MorfologikSlovenianSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
	}
}

// GetRelevantRuleIDs ports Slovenian.getRelevantRules on SlovenianLang.
func (v SlovenianLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), SlovenianRelevantRuleIDs()...)
}
