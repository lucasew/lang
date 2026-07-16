package language

// Additional SmallLang entries not in the first batch.
var (
Khmer      = SmallLang{"km", "Khmer", "HUNSPELL_RULE_KM", []string{"KH"}}
Malayalam  = SmallLang{"ml", "Malayalam", "HUNSPELL_RULE_ML", []string{"IN"}}
Tagalog    = SmallLang{"tl", "Tagalog", "MORFOLOGIK_RULE_TL", []string{"PH"}}
Tamil      = SmallLang{"ta", "Tamil", "HUNSPELL_RULE_TA", []string{"IN", "LK"}}
Lithuanian = SmallLang{"lt", "Lithuanian", "MORFOLOGIK_RULE_LT", []string{"LT"}}
Icelandic  = SmallLang{"is", "Icelandic", "HUNSPELL_RULE_IS", []string{"IS"}}
Belarusian = SmallLang{"be", "Belarusian", "MORFOLOGIK_RULE_BE", []string{"BY"}}
Breton     = SmallLang{"br", "Breton", "MORFOLOGIK_RULE_BR", []string{"FR"}}
CrimeanTatar = SmallLang{"crh", "Crimean Tatar", "MORFOLOGIK_RULE_CRH", []string{"UA"}}
Slovenian  = SmallLang{"sl", "Slovenian", "MORFOLOGIK_RULE_SL", []string{"SI"}}
Asturian   = SmallLang{"ast", "Asturian", "MORFOLOGIK_RULE_AST", []string{"ES"}}
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
