package language

// Language-model-dependent rule ID surfaces (getRelevantLanguageModelRules /
// getRelevantLanguageModelCapableRules). IDs from Java getId / RULE_ID only.

// GermanLanguageModelRelevantRuleIDs ports German.getRelevantLanguageModelRules.
func GermanLanguageModelRelevantRuleIDs() []string {
	return []string{
		"DE_UPPER_CASE_NGRAM",
		"CONFUSION_RULE", // GermanConfusionProbabilityRule extends ConfusionProbabilityRule
		"DE_PROHIBITED_COMPOUNDS",
	}
}

// GermanyGermanLanguageModelCapableRuleIDs ports
// GermanyGerman.getRelevantLanguageModelCapableRules extras (speller when configured).
// Base Language.getRelevantLanguageModelCapableRules is empty; variants add spellers.
func GermanyGermanLanguageModelCapableRuleIDs() []string {
	return []string{"GERMAN_SPELLER_RULE"}
}

// AustrianGermanLanguageModelCapableRuleIDs ports AustrianGerman capable speller.
func AustrianGermanLanguageModelCapableRuleIDs() []string {
	return []string{"AUSTRIAN_GERMAN_SPELLER_RULE"}
}

// SwissGermanLanguageModelCapableRuleIDs ports SwissGerman capable speller.
func SwissGermanLanguageModelCapableRuleIDs() []string {
	return []string{"SWISS_GERMAN_SPELLER_RULE"}
}

// GermanVariantLanguageModelCapableRuleIDs selects DE/AT/CH capable spellers.
func GermanVariantLanguageModelCapableRuleIDs(v GermanVariant) []string {
	switch {
	case isSwissGermanVariant(v):
		return SwissGermanLanguageModelCapableRuleIDs()
	case equalFoldASCII(v.ShortCode, "de-AT"):
		return AustrianGermanLanguageModelCapableRuleIDs()
	default:
		// de-DE and NonSwissGerman surface use GermanyGerman speller
		return GermanyGermanLanguageModelCapableRuleIDs()
	}
}

// EnglishLanguageModelRelevantRuleIDs ports English.getRelevantLanguageModelRules.
func EnglishLanguageModelRelevantRuleIDs() []string {
	return []string{
		"EN_UPPER_CASE_NGRAM",
		"CONFUSION_RULE", // EnglishConfusionProbabilityRule
		"NGRAM_RULE",     // EnglishNgramProbabilityRule
	}
}

// EnglishLanguageModelCapableRuleIDsForMotherTongue ports
// English.getRelevantLanguageModelCapableRules when lm != nil and motherTongue matches.
// Empty motherTongue short code → empty list (Java emptyList).
func EnglishLanguageModelCapableRuleIDsForMotherTongue(motherShortCode string) []string {
	switch motherShortCode {
	case "fr":
		return []string{"EN_FOR_FR_SPEAKERS_FALSE_FRIENDS"}
	case "de":
		return []string{"EN_FOR_DE_SPEAKERS_FALSE_FRIENDS"}
	case "es":
		return []string{"EN_FOR_ES_SPEAKERS_FALSE_FRIENDS"}
	case "nl":
		return []string{"EN_FOR_NL_SPEAKERS_FALSE_FRIENDS"}
	default:
		return nil
	}
}

// EnglishVariantLanguageModelCapableSpellerRuleIDs ports locale speller added in
// getRelevantLanguageModelCapableRules (default path; SymSpell experiment omitted).
// Combined with EnglishLanguageModelCapableRuleIDsForMotherTongue when L1 is set.
func EnglishVariantLanguageModelCapableSpellerRuleIDs(v EnglishVariant) []string {
	// Java: Morfologik*SpellerRule for each locale (AmericanEnglish may use SymSpell
	// only under SuggestionsChanges experiment — not invent experiment path).
	if v.SpellerRuleID != "" {
		return []string{v.SpellerRuleID}
	}
	return nil
}

// FrenchLanguageModelRelevantRuleIDs ports French.getRelevantLanguageModelRules.
func FrenchLanguageModelRelevantRuleIDs() []string {
	return []string{"CONFUSION_RULE"} // FrenchConfusionProbabilityRule
}

// PortugueseLanguageModelRelevantRuleIDs ports Portuguese.getRelevantLanguageModelRules.
func PortugueseLanguageModelRelevantRuleIDs() []string {
	return []string{"CONFUSION_RULE"} // PortugueseConfusionProbabilityRule
}

// SpanishLanguageModelRelevantRuleIDs ports Spanish.getRelevantLanguageModelRules.
func SpanishLanguageModelRelevantRuleIDs() []string {
	return []string{"CONFUSION_RULE"} // SpanishConfusionProbabilityRule
}

// RussianLanguageModelRelevantRuleIDs ports Russian.getRelevantLanguageModelRules.
func RussianLanguageModelRelevantRuleIDs() []string {
	return []string{"CONFUSION_RULE"} // RussianConfusionProbabilityRule
}

// ItalianLanguageModelRelevantRuleIDs ports Italian.getRelevantLanguageModelRules.
func ItalianLanguageModelRelevantRuleIDs() []string {
	return []string{"CONFUSION_RULE"} // ItalianConfusionProbabilityRule
}

// DutchLanguageModelRelevantRuleIDs ports Dutch.getRelevantLanguageModelRules
// (commented-out confusion rule in Java — method exists but body returns empty?
// Re-check: Dutch has commented block for confusion; if method is commented, no surface.
// Current Dutch.java has the method commented out entirely in source earlier — check).
func DutchLanguageModelRelevantRuleIDs() []string {
	// Java Dutch.java: getRelevantLanguageModelRules is fully commented out
	// (// commented out as long as there are not enough entries…).
	// Faithful: no active LM rule list.
	return nil
}

// ArabicLanguageModelRelevantRuleIDs ports Arabic.getRelevantLanguageModelRules.
func ArabicLanguageModelRelevantRuleIDs() []string {
	return []string{"CONFUSION_RULE"} // ArabicConfusionProbabilityRule
}

// ChineseLanguageModelRelevantRuleIDs ports Chinese.getRelevantLanguageModelRules.
func ChineseLanguageModelRelevantRuleIDs() []string {
	return []string{"CONFUSION_RULE"} // ChineseConfusionProbabilityRule
}

// SimpleGermanLanguageModelRelevantRuleIDs ports SimpleGerman.getRelevantLanguageModelRules → empty.
func SimpleGermanLanguageModelRelevantRuleIDs() []string { return nil }

// SimpleGermanLanguageModelCapableRuleIDs ports SimpleGerman capable → empty.
func SimpleGermanLanguageModelCapableRuleIDs() []string { return nil }
