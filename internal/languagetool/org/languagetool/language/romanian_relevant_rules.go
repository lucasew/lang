package language

// RomanianRelevantRuleIDs ports Romanian.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Romanian.java.
func RomanianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",        // CommaWhitespaceRule
		"DOUBLE_PUNCTUATION",                  // DoublePunctuationRule
		"UPPERCASE_SENTENCE_START",            // UppercaseSentenceStartRule
		"WHITESPACE_RULE",                     // MultipleWhitespaceRule
		"UNPAIRED_BRACKETS",                   // GenericUnpairedBracketsRule (RO symbols)
		"WORD_REPEAT_RULE",                    // WordRepeatRule
		"MORFOLOGIK_RULE_RO_RO",               // MorfologikRomanianSpellerRule
		"ROMANIAN_WORD_REPEAT_BEGINNING_RULE", // RomanianWordRepeatBeginningRule
		"RO_SIMPLE_REPLACE",                   // SimpleReplaceRule
		"RO_COMPOUND",                         // CompoundRule
	}
}

// GetRelevantRuleIDs ports Romanian.getRelevantRules on RomanianLang.
func (v RomanianLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), RomanianRelevantRuleIDs()...)
}
