package language

// CrimeanTatarRelevantRuleIDs ports CrimeanTatar.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in CrimeanTatar.java exactly (8 rules, no SharedLayout invent).
func CrimeanTatarRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"WHITESPACE_PARAGRAPH",         // WhiteSpaceBeforeParagraphEnd
		"WHITESPACE_PARAGRAPH_BEGIN",   // WhiteSpaceAtBeginOfParagraph
		"MORFOLOGIK_RULE_CRH_UA",       // MorfologikCrimeanTatarSpellerRule
	}
}

// GetRelevantRuleIDs ports CrimeanTatar.getRelevantRules on CrimeanTatarLang.
func (v CrimeanTatarLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), CrimeanTatarRelevantRuleIDs()...)
}
