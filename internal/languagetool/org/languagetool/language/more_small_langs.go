package language

// Additional SmallLang entries not in the first batch.
// SpellerRuleID = Java createDefaultSpellingRule / registered speller getId only.
var (
	// KhmerHunspellRule extends HunspellRule without getId override → HUNSPELL_RULE
	Khmer = SmallLang{"km", "Khmer", "HUNSPELL_RULE", []string{"KH"}}
	// MorfologikMalayalamSpellerRule.getId
	Malayalam = SmallLang{"ml", "Malayalam", "MORFOLOGIK_RULE_ML_IN", []string{"IN"}}
	// Tagalog promoted to TagalogLang (see tagalog.go) — not SmallLang.
	// Tamil: no default speller in Java Language module
	Tamil      = SmallLang{"ta", "Tamil", "", []string{"IN", "LK"}}
	Lithuanian = SmallLang{"lt", "Lithuanian", "MORFOLOGIK_RULE_LT_LT", []string{"LT"}}
	// HunspellNoSuggestionRule.RULE_ID
	Icelandic = SmallLang{"is", "Icelandic", "HUNSPELL_NO_SUGGEST_RULE", []string{"IS"}}
	// Belarusian promoted to BelarusianLang (see belarusian.go) — not SmallLang.
	// Breton promoted to BretonLang (see breton.go) — not SmallLang.
	// MorfologikCrimeanTatarSpellerRule.getId
	CrimeanTatar = SmallLang{"crh", "Crimean Tatar", "MORFOLOGIK_RULE_CRH_UA", []string{"UA"}}
	// Slovenian promoted to SlovenianLang (see slovenian.go) — not SmallLang.
	// Asturian promoted to AsturianLang (see asturian.go) — not SmallLang.
)

// AllExtendedSmallLangs returns the original set plus additional modules.
func AllExtendedSmallLangs() []SmallLang {
	base := AllSmallLangs()
	extra := []SmallLang{
		Khmer, Malayalam, Tamil, Lithuanian, Icelandic,
		CrimeanTatar,
	}
	return append(base, extra...)
}
