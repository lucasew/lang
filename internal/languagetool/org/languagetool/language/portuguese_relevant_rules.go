package language

// PortugueseRelevantRuleIDsBase ports Portuguese.getRelevantRules class IDs.
// MorfologikPortugueseSpellerRule ID is "MORFOLOGIK_RULE_" +
// shortCodeWithCountryAndVariant (e.g. pt-PT → MORFOLOGIK_RULE_PT_PT); variants
// inject SpellerRuleID via GetRelevantRuleIDs.
// Order matches Java Arrays.asList in Portuguese.java (speller slot marked).
func PortugueseRelevantRuleIDsBase(spellerRuleID string) []string {
	if spellerRuleID == "" {
		spellerRuleID = "MORFOLOGIK_RULE_PT_PT" // PortugalPortuguese default surface
	}
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"UNPAIRED_BRACKETS",
		spellerRuleID,
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"EMPTY_LINE",
		"PARAGRAPH_REPEAT_BEGINNING_RULE",
		"PUNCTUATION_PARAGRAPH_END",
		"PT_COMPOUNDS_POST_REFORM",
		"PT_COLOUR_HYPHENATION",
		"PT_SIMPLE_REPLACE_ORTHOGRAPHY",
		"PT_SIMPLE_REPLACE",
		"PT_BARBARISMS_REPLACE",
		// PortugueseArchaismsRule is commented out in Java getRelevantRules (#3095) — omit.
		"PT_CLICHE_REPLACE",
		"FILLER_WORDS_PT",
		"PT_REDUNDANCY_REPLACE",
		"PT_WORDINESS_REPLACE",
		// PortugueseWeaselWordsRule is commented out in Java getRelevantRules — omit.
		"PT_WIKIPEDIA_COMMON_ERRORS",
		"PORTUGUESE_WORD_REPEAT_RULE",
		"PORTUGUESE_WORD_REPEAT_BEGINNING_RULE",
		"ACCENTUATION_CHECK_PT",
		"PT_DIACRITICS_REPLACE",
		"PORTUGUESE_WRONG_WORD_IN_CONTEXT",
		"PT_WORD_COHERENCY",
		"UNIDADES_METRICAS",
		// PortugueseReadabilityRule(true) then (false)
		"READABILITY_RULE_SIMPLE_PT",
		"READABILITY_RULE_DIFFICULT_PT",
		"DOUBLE_PUNCTUATION",
		"PT_ENGLISH_CONTRACTION_ORTHOGRAPHY",
	}
}

// PortugueseRelevantRuleIDs is the Portugal (pt-PT) surface of the base list.
func PortugueseRelevantRuleIDs() []string {
	return PortugueseRelevantRuleIDsBase("MORFOLOGIK_RULE_PT_PT")
}

// GetRelevantRuleIDs ports Portuguese.getRelevantRules for Portuguese variants.
func (v PortugueseVariant) GetRelevantRuleIDs() []string {
	return PortugueseRelevantRuleIDsBase(v.SpellerRuleID)
}
