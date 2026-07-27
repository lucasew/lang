package language

// BelarusianRelevantRuleIDs ports Belarusian.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Belarusian.java exactly (14 rules).
// Note: Java comments out PunctuationMarkAtParagraphEnd; PunctuationMarkAtParagraphEnd2 is kept.
// Java BelarusianSpecificCaseRule.getId() returns BE_SPECIFIC_CASE (not BELARUSIAN_SPECIFIC_CASE;
// the priority switch key BELARUSIAN_SPECIFIC_CASE is a separate Java id — see belarusian_priority.go).
func BelarusianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",           // DoublePunctuationRule
		"MORFOLOGIK_RULE_BE_BY",        // MorfologikBelarusianSpellerRule
		"UPPERCASE_SENTENCE_START",     // UppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"WHITESPACE_PARAGRAPH",         // WhiteSpaceBeforeParagraphEnd
		"WHITESPACE_PARAGRAPH_BEGIN",   // WhiteSpaceAtBeginOfParagraph
		"TOO_LONG_SENTENCE",            // LongSentenceRule(…, 50)
		"TOO_LONG_PARAGRAPH",           // LongParagraphRule
		"PARAGRAPH_REPEAT_BEGINNING_RULE", // ParagraphRepeatBeginningRule
		// PunctuationMarkAtParagraphEnd commented out in Java Belarusian.java
		"PUNCTUATION_PARAGRAPH_END2", // PunctuationMarkAtParagraphEnd2
		"BE_SIMPLE_REPLACE",          // SimpleReplaceRule
		"BE_SPECIFIC_CASE",           // BelarusianSpecificCaseRule.getId
	}
}

// GetRelevantRuleIDs ports Belarusian.getRelevantRules on BelarusianLang.
func (v BelarusianLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), BelarusianRelevantRuleIDs()...)
}
