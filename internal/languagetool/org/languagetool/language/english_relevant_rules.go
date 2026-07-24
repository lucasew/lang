package language

// EnglishRelevantRuleIDs ports English.getRelevantRules base list (class getId only).
// OpenNMTRule is commented out in Java (#903) — omitted.
// Mother-tongue L2 grammar XML loads are not listed (pattern-file, not rule class IDs).
// Order matches English.java asList after optional L2 grammar.
func EnglishRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"EMPTY_LINE",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"PARAGRAPH_REPEAT_BEGINNING_RULE",
		"PUNCTUATION_PARAGRAPH_END",
		"PUNCTUATION_PARAGRAPH_END2",
		"EN_CONSISTENT_APOS",
		"EN_SPECIFIC_CASE",
		"EN_UNPAIRED_BRACKETS",
		"EN_UNPAIRED_QUOTES",
		"ENGLISH_WORD_REPEAT_RULE",
		"EN_A_VS_AN",
		"ENGLISH_WORD_REPEAT_BEGINNING_RULE",
		"EN_COMPOUNDS",
		"EN_CONTRACTION_SPELLING",
		"ENGLISH_WRONG_WORD_IN_CONTEXT",
		"EN_DASH_RULE",
		"EN_WORD_COHERENCY",
		"EN_DIACRITICS_REPLACE_ORTHOGRAPHY",
		"EN_PLAIN_ENGLISH_REPLACE",
		"EN_REDUNDANCY_REPLACE",
		"EN_SIMPLE_REPLACE",
		"PROFANITY",
		// ReadabilityRule(false) then (true)
		"READABILITY_RULE_DIFFICULT",
		"READABILITY_RULE_SIMPLE",
		"EN_REPEATEDWORDS",
		"TOO_OFTEN_USED_VERB_EN",
		"TOO_OFTEN_USED_NOUN_EN",
		"TOO_OFTEN_USED_ADJECTIVE_EN",
	}
}

// GetRelevantRuleIDs ports English.getRelevantRules + locale extras.
func (v EnglishVariant) GetRelevantRuleIDs() []string {
	out := append([]string(nil), EnglishRelevantRuleIDs()...)
	out = append(out, v.RelevantExtraRuleIDs...)
	return out
}
