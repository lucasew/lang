package language

// GreekRelevantRuleIDs ports Greek.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Greek.java.
func GreekRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		// Custom unpaired id (not default UNPAIRED_BRACKETS)
		"EL_UNPAIRED_BRACKETS",
		"TOO_LONG_SENTENCE",
		"MORFOLOGIK_RULE_EL_GR",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"GREEK_WORD_REPEAT_BEGINNING_RULE",
		"WORD_REPEAT_RULE",
		"GREEK_HOMONYMS_REPLACE",
		"EL_SPECIFIC_CASE",
		"GREEK_ORTHOGRAPHY_NUMERAL_STRESS",
		"EL_REDUNDANCY_REPLACE",
	}
}
