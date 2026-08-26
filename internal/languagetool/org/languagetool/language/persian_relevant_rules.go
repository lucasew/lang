package language

// PersianRelevantRuleIDs ports Persian.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Persian.java exactly (11 rules).
// No SharedLayout invent (no uppercase/unpaired/empty-line unless Java lists them).
// No speller (Java has none in getRelevantRules).
func PersianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",        // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",                  // DoublePunctuationRule
		"WHITESPACE_RULE",                     // MultipleWhitespaceRule
		"TOO_LONG_SENTENCE",                   // LongSentenceRule(…, 50)
		"PERSIAN_COMMA_PARENTHESIS_WHITESPACE", // PersianCommaWhitespaceRule
		"PERSIAN_DOUBLE_PUNCTUATION",          // PersianDoublePunctuationRule
		"PERSIAN_WORD_REPEAT_BEGINNING_RULE",  // PersianWordRepeatBeginningRule
		"PERSIAN_WORD_REPEAT_RULE",            // PersianWordRepeatRule
		"FA_SIMPLE_REPLACE",                   // SimpleReplaceRule
		"FA_SPACE_BEFORE_CONJUNCTION",         // PersianSpaceBeforeRule
		"FA_WORD_COHERENCY",                   // WordCoherencyRule
	}
}

// GetRelevantRuleIDs ports Persian.getRelevantRules on PersianLang.
func (v PersianLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), PersianRelevantRuleIDs()...)
}
