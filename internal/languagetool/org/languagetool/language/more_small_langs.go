package language

// Additional SmallLang entries not in the first batch.
// SpellerRuleID = Java createDefaultSpellingRule / registered speller getId only.
var (
	// KhmerHunspellRule extends HunspellRule without getId override → HUNSPELL_RULE
	Khmer = SmallLang{"km", "Khmer", "HUNSPELL_RULE", []string{"KH"}}
	// MorfologikMalayalamSpellerRule.getId
	Malayalam = SmallLang{"ml", "Malayalam", "MORFOLOGIK_RULE_ML_IN", []string{"IN"}}
	Tagalog   = SmallLang{"tl", "Tagalog", "MORFOLOGIK_RULE_TL", []string{"PH"}}
	// Tamil: no default speller in Java Language module
	Tamil      = SmallLang{"ta", "Tamil", "", []string{"IN", "LK"}}
	Lithuanian = SmallLang{"lt", "Lithuanian", "MORFOLOGIK_RULE_LT_LT", []string{"LT"}}
	// HunspellNoSuggestionRule.RULE_ID
	Icelandic = SmallLang{"is", "Icelandic", "HUNSPELL_NO_SUGGEST_RULE", []string{"IS"}}
	// MorfologikBelarusianSpellerRule.getId
	Belarusian = SmallLang{"be", "Belarusian", "MORFOLOGIK_RULE_BE_BY", []string{"BY"}}
	// MorfologikBretonSpellerRule.getId
	Breton = SmallLang{"br", "Breton", "MORFOLOGIK_RULE_BR_FR", []string{"FR"}}
	// MorfologikCrimeanTatarSpellerRule.getId
	CrimeanTatar = SmallLang{"crh", "Crimean Tatar", "MORFOLOGIK_RULE_CRH_UA", []string{"UA"}}
	// MorfologikSlovenianSpellerRule.getId
	Slovenian = SmallLang{"sl", "Slovenian", "MORFOLOGIK_RULE_SL_SI", []string{"SI"}}
	Asturian  = SmallLang{"ast", "Asturian", "MORFOLOGIK_RULE_AST", []string{"ES"}}
)

// AllExtendedSmallLangs returns the original set plus additional modules.
func AllExtendedSmallLangs() []SmallLang {
base := AllSmallLangs()
extra := []SmallLang{
	Khmer, Malayalam, Tagalog, Tamil, Lithuanian, Icelandic,
	Belarusian, Breton, CrimeanTatar, Slovenian, Asturian,
}
return append(base, extra...)
}
