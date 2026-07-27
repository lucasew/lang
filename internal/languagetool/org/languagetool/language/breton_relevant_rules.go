package language

// BretonRelevantRuleIDs ports Breton.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Breton.java exactly.
// Note: BretonCompoundRule (BR_COMPOUNDS) exists in rules/br but is NOT in Java
// getRelevantRules — do not invent into this list.
func BretonRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"MORFOLOGIK_RULE_BR_FR",        // MorfologikBretonSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"BR_TOPO",                      // TopoReplaceRule
	}
}

// GetRelevantRuleIDs ports Breton.getRelevantRules on BretonLang.
func (v BretonLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), BretonRelevantRuleIDs()...)
}
