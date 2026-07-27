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

// PortugueseRelevantRuleIDs ports base Portuguese.getRelevantRules with pt-PT speller.
// Does not include PortugalPortuguese extras (see PortugalPortugueseRelevantRuleIDs).
func PortugueseRelevantRuleIDs() []string {
	return PortugueseRelevantRuleIDsBase("MORFOLOGIK_RULE_PT_PT")
}

// PortugalPortugueseRelevantExtraRuleIDs ports the rules PortugalPortuguese adds
// after super.getRelevantRules (Java PortugalPortuguese.java order).
// Same getId as a base rule appears again when Java constructs a second instance
// (PostReformPortugueseCompoundRule; path packs for barbarisms/cliche/…).
func PortugalPortugueseRelevantExtraRuleIDs() []string {
	return []string{
		"PT_COMPOUNDS_POST_REFORM",   // second PostReformPortugueseCompoundRule
		"PT_POSAO_DASH_RULE",         // PostReformPortugueseDashRule
		"PT_AGREEMENT_REPLACE",       // PortugueseAgreementReplaceRule
		"PT_PT_SIMPLE_REPLACE",       // PortugalPortugueseReplaceRule("/pt/pt-PT/replace.txt")
		"PT_BARBARISMS_REPLACE",      // PortugueseBarbarismsRule("/pt/pt-PT/barbarisms.txt")
		"PT_ARCHAISMS_REPLACE",       // PortugueseArchaismsRule("/pt/pt-PT/archaisms.txt")
		"PT_CLICHE_REPLACE",          // PortugueseClicheRule("/pt/pt-PT/cliches.txt")
		"PT_REDUNDANCY_REPLACE",      // PortugueseRedundancyRule("/pt/pt-PT/redundancies.txt")
		"PT_WORDINESS_REPLACE",       // PortugueseWordinessRule("/pt/pt-PT/wordiness.txt")
		"PT_WIKIPEDIA_COMMON_ERRORS", // PortugueseWikipediaRule("/pt/pt-PT/wikipedia.txt")
	}
}

// PortugalPortugueseRelevantRuleIDs ports PortugalPortuguese.getRelevantRules
// = super + extras in Java ArrayList order.
func PortugalPortugueseRelevantRuleIDs() []string {
	out := PortugueseRelevantRuleIDsBase("MORFOLOGIK_RULE_PT_PT")
	return append(out, PortugalPortugueseRelevantExtraRuleIDs()...)
}

// BrazilianPortugueseRelevantExtraRuleIDs ports the rules BrazilianPortuguese adds
// after super.getRelevantRules (Java BrazilianPortuguese.java order).
// Differs from PortugalPortuguese: BrazilianPortugueseReplaceRule instead of
// PortugalPortugueseReplaceRule; no PortugueseAgreementReplaceRule.
// Same getId as a base rule appears again when Java constructs a second instance
// (PostReformPortugueseCompoundRule; path packs for barbarisms/cliche/…).
func BrazilianPortugueseRelevantExtraRuleIDs() []string {
	return []string{
		"PT_COMPOUNDS_POST_REFORM",   // second PostReformPortugueseCompoundRule
		"PT_POSAO_DASH_RULE",         // PostReformPortugueseDashRule
		"PT_BR_SIMPLE_REPLACE",       // BrazilianPortugueseReplaceRule("/pt/pt-BR/replace.txt")
		"PT_BARBARISMS_REPLACE",      // PortugueseBarbarismsRule("/pt/pt-BR/barbarisms.txt")
		"PT_ARCHAISMS_REPLACE",       // PortugueseArchaismsRule("/pt/pt-BR/archaisms.txt")
		"PT_CLICHE_REPLACE",          // PortugueseClicheRule("/pt/pt-BR/cliches.txt")
		"PT_REDUNDANCY_REPLACE",      // PortugueseRedundancyRule("/pt/pt-BR/redundancies.txt")
		"PT_WORDINESS_REPLACE",       // PortugueseWordinessRule("/pt/pt-BR/wordiness.txt")
		"PT_WIKIPEDIA_COMMON_ERRORS", // PortugueseWikipediaRule("/pt/pt-BR/wikipedia.txt")
	}
}

// BrazilianPortugueseRelevantRuleIDs ports BrazilianPortuguese.getRelevantRules
// = super + extras in Java ArrayList order (speller MORFOLOGIK_RULE_PT_BR).
func BrazilianPortugueseRelevantRuleIDs() []string {
	out := PortugueseRelevantRuleIDsBase("MORFOLOGIK_RULE_PT_BR")
	return append(out, BrazilianPortugueseRelevantExtraRuleIDs()...)
}

// AngolaPortugueseRelevantExtraRuleIDs ports the rules AngolaPortuguese adds
// after super.getRelevantRules (Java AngolaPortuguese.java order).
// Pre-reform only: PreReformPortugueseCompoundRule + PreReformPortugueseDashRule.
// No Portugal/BR path packs, no agreement, no PostReform extras, no priority override.
func AngolaPortugueseRelevantExtraRuleIDs() []string {
	return []string{
		"PT_COMPOUNDS_PRE_REFORM", // PreReformPortugueseCompoundRule
		"PT_PREAO_DASH_RULE",      // PreReformPortugueseDashRule
	}
}

// AngolaPortugueseRelevantRuleIDs ports AngolaPortuguese.getRelevantRules
// = super + extras in Java ArrayList order (speller MORFOLOGIK_RULE_PT_AO).
func AngolaPortugueseRelevantRuleIDs() []string {
	out := PortugueseRelevantRuleIDsBase("MORFOLOGIK_RULE_PT_AO")
	return append(out, AngolaPortugueseRelevantExtraRuleIDs()...)
}

// GetRelevantRuleIDs ports Portuguese / PortugalPortuguese / BrazilianPortuguese /
// AngolaPortuguese getRelevantRules.
// PortugalPortuguese (pt-PT, bare pt product surface), BrazilianPortuguese (pt-BR),
// and AngolaPortuguese (pt-AO) append their variant extras after base.
func (v PortugueseVariant) GetRelevantRuleIDs() []string {
	out := PortugueseRelevantRuleIDsBase(v.SpellerRuleID)
	if isPortugalPortugueseCode(v.ShortCode) {
		out = append(out, PortugalPortugueseRelevantExtraRuleIDs()...)
	} else if isBrazilianPortugueseCode(v.ShortCode) {
		out = append(out, BrazilianPortugueseRelevantExtraRuleIDs()...)
	} else if isAngolaPortugueseCode(v.ShortCode) {
		out = append(out, AngolaPortugueseRelevantExtraRuleIDs()...)
	}
	return out
}
