package language

// EsperantoRelevantRuleIDs ports Esperanto.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Esperanto.java exactly (8 rules).
// GenericUnpairedBracketsRule uses default pairs; id remains UNPAIRED_BRACKETS.
// HunspellRule id HUNSPELL_RULE (also createDefaultSpellingRule).
// DateCheckFilter is NOT in getRelevantRules — do not invent.
// No SharedLayout invent (no empty line, long paragraph, paragraph-repeat, …).
func EsperantoRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule (default pairs)
		"HUNSPELL_RULE",                // HunspellRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WORD_REPEAT_RULE",             // WordRepeatRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
	}
}

// GetRelevantRuleIDs ports Esperanto.getRelevantRules on EsperantoLang.
func (v EsperantoLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), EsperantoRelevantRuleIDs()...)
}
