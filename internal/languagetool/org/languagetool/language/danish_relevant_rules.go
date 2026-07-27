package language

// DanishRelevantRuleIDs ports Danish.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Danish.java exactly (6 rules).
// GenericUnpairedBracketsRule uses custom start/end symbols; id remains UNPAIRED_BRACKETS.
// WORD_REPEAT_RULE is in grammar.xml (Java comment) — not class-based.
func DanishRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (custom start/end)
		"HUNSPELL_RULE",                // HunspellRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		// "WORD_REPEAT_RULE" implemented in grammar.xml
		"WHITESPACE_RULE", // MultipleWhitespaceRule
	}
}

// GetRelevantRuleIDs ports Danish.getRelevantRules on DanishLang.
func (v DanishLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), DanishRelevantRuleIDs()...)
}
