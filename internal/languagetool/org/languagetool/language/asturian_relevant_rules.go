package language

// AsturianRelevantRuleIDs ports Asturian.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Asturian.java exactly.
func AsturianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (default symbols)
		"MORFOLOGIK_RULE_AST",          // MorfologikAsturianSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
	}
}

// GetRelevantRuleIDs ports Asturian.getRelevantRules on AsturianLang.
func (v AsturianLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), AsturianRelevantRuleIDs()...)
}
