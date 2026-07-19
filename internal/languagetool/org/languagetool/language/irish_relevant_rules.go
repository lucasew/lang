package language

// IrishRelevantRuleIDs ports Irish.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Irish.java — UppercaseSentenceStartRule
// appears twice in the live Java list (duplicate registration).
func IrishRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"UNPAIRED_BRACKETS",
		"DOUBLE_PUNCTUATION",
		"UPPERCASE_SENTENCE_START",
		"TOO_LONG_SENTENCE",
		"TOO_LONG_PARAGRAPH",
		"UPPERCASE_SENTENCE_START", // second registration in Java asList
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WHITESPACE_PARAGRAPH",
		"WHITESPACE_PARAGRAPH_BEGIN",
		"PARAGRAPH_REPEAT_BEGINNING_RULE",
		"WORD_REPEAT_RULE",
		"MORFOLOGIK_RULE_GA_IE",
		"GA_LOGAINM",
		"GA_PEOPLE",
		"GA_SPASANNA",
		"GA_COMPOUNDS",
		"GA_PRESTANDARD_REPLACE",
		"GA_REPLACE",
		"GA_FGB_EQ_REPLACE",
		"GA_ENGLISH_HOMOPHONE",
		"GA_DHA_NO_BEIRT",
		"GA_DATIVE_PLURALS_STD",
		"GA_SPECIFIC_CASE",
	}
}
