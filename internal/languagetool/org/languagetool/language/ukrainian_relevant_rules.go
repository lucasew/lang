package language

// UkrainianRelevantRuleIDs ports Ukrainian.getRelevantRules rule IDs (class getId only).
// Order matches Java Arrays.asList in Ukrainian.java (live new … entries only;
// commented DoublePunctuationRule omitted).
func UkrainianRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE", // UkrainianCommaWhitespaceRule → CommaWhitespaceRule
		"UPPERCASE_SENTENCE_START",     // UkrainianUppercaseSentenceStartRule
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"UKRAINIAN_WORD_REPEAT_RULE",
		"DASH", // TypographyRule
		"UK_HIDDEN_CHARS",
		"MORFOLOGIK_RULE_UK_UA",
		"UK_MISSING_HYPHEN",
		"UK_VERB_NOUN_INFLECTION_AGREEMENT",
		"UK_NOUN_VERB_INFLECTION_AGREEMENT",
		"UK_ADJ_NOUN_INFLECTION_AGREEMENT",
		"UK_PREP_NOUN_INFLECTION_AGREEMENT",
		"UK_NUMR_NOUN_INFLECTION_AGREEMENT",
		"UK_MIXED_ALPHABETS",
		"UK_SIMPLE_REPLACE_SOFT",
		"UK_SIMPLE_REPLACE_RENAMED",
		"UK_SIMPLE_REPLACE",
	}
}
