package language

// RomanianRelevantRuleIDs ports Romanian.getRelevantRules rule IDs (class getId only).
func RomanianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"UNPAIRED_BRACKETS",
		"WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_RO_RO",
		"ROMANIAN_WORD_REPEAT_BEGINNING_RULE",
		"RO_SIMPLE_REPLACE",
		"RO_COMPOUND",
	}
}
