package language

// RussianRelevantRuleIDs ports Russian.getRelevantRules rule IDs (class getId only).
// Order matches live new … entries in Russian.java (commented-out rules omitted).
func RussianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		// DoublePunctuationRule commented out in Java (XML rule instead)
		"UPPERCASE_SENTENCE_START",
		"MORFOLOGIK_RULE_RU_RU",
		// WordRepeatRule commented out — moved to RussianSimpleWordRepeatRule
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		// EmptyLineRule commented out
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE",
		"FILLER_WORDS_RU",
		// PunctuationMarkAtParagraphEnd commented out
		"PUNCTUATION_PARAGRAPH_END2",
		// ReadabilityRule pair commented out
		// specific to Russian:
		"MORFOLOGIK_RULE_RU_RU_YO",
		"RU_UNPAIRED_BRACKETS",
		"RU_COMPOUNDS",
		"RU_SIMPLE_REPLACE",
		"WORD_REPEAT_RULE", // RussianSimpleWordRepeatRule extends WordRepeatRule
		"RU_WORD_COHERENCY",
		"RU_WORD_REPEAT",
		"RU_WORD_ROOT_REPEAT",
		"RU_VERB_CONJUGATION",
		"RU_DASH_RULE",
		"RU_SPECIFIC_CASE",
	}
}

// GetRelevantRuleIDs ports Russian.getRelevantRules on RussianLang.
func (v RussianLang) GetRelevantRuleIDs() []string {
	return append([]string(nil), RussianRelevantRuleIDs()...)
}
