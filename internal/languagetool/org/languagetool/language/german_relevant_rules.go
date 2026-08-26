package language

// GermanRelevantRuleIDs ports German.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in German.java. Language-model / speller rules
// are getRelevantLanguageModel* — not listed here.
func GermanRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // GermanCommaWhitespaceRule → CommaWhitespaceRule
		"UNPAIRED_BRACKETS",            // GermanUnpairedBracketsRule
		"DE_UNPAIRED_QUOTES",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"EMPTY_LINE",
		"TOO_LONG_PARAGRAPH",
		"PUNCTUATION_PARAGRAPH_END",
		"DE_SIMPLE_REPLACE",
		"OLD_SPELLING_RULE",
		"DE_SENTENCE_WHITESPACE",
		"DE_DOUBLE_PUNCTUATION",
		"MISSING_VERB",
		"GERMAN_WORD_REPEAT_RULE",
		"GERMAN_WORD_REPEAT_BEGINNING_RULE",
		"GERMAN_WRONG_WORD_IN_CONTEXT",
		"DE_AGREEMENT",
		"DE_AGREEMENT2",
		"DE_CASE",
		"DE_DASH",
		"DE_VERBAGREEMENT",
		"DE_SUBJECT_VERB_AGREEMENT",
		"DE_WORD_COHERENCY",
		"DE_SIMILAR_NAMES",
		"DE_WIEDER_VS_WIDER",
		"STYLE_REPEATED_WORD_RULE_DE",
		"DE_COMPOUND_COHERENCY",
		"TOO_LONG_SENTENCE_DE",
		"FILLER_WORDS_DE",
		"PASSIVE_SENTENCE_DE",
		"SENTENCE_WITH_MODAL_VERB_DE",
		"SENTENCE_WITH_MAN_DE",
		"SENTENCE_BEGINNING_WITH_CONJUNCTION_DE",
		"NON_SIGNIFICANT_VERB_DE",
		"UNNECESSARY_PHRASES_DE",
		"GERMAN_PARAGRAPH_REPEAT_BEGINNING_RULE",
		"DE_DU_UPPER_LOWER",
		"EINHEITEN_METRISCH",
		// MissingCommaRelativeClauseRule(false) then (true)
		"COMMA_IN_FRONT_RELATIVE_CLAUSE",
		"COMMA_BEHIND_RELATIVE_CLAUSE",
		"REDUNDANT_MODAL_VERB",
		// GermanReadabilityRule(true) then (false)
		"READABILITY_RULE_SIMPLE_DE",
		"READABILITY_RULE_DIFFICULT_DE",
		"COMPOUND_INFINITIV_RULE",
		"STYLE_REPEATED_SHORT_SENTENCES",
		"STYLE_REPEATED_SENTENCE_BEGINNING",
		"DE_REPEATEDWORDS", // AbstractRepeatedWordsRule: shortCode_REPEATEDWORDS
		"TOO_OFTEN_USED_VERB_DE",
		"TOO_OFTEN_USED_NOUN_DE",
		"TOO_OFTEN_USED_ADJECTIVE_DE",
	}
}

// GetRelevantRuleIDs ports German.getRelevantRules (+ DE/AT DE_COMPOUNDS extras).
func (v GermanVariant) GetRelevantRuleIDs() []string {
	out := append([]string(nil), GermanRelevantRuleIDs()...)
	out = append(out, v.RelevantExtraRuleIDs...)
	return out
}
