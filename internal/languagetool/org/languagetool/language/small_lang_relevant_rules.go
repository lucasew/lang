package language

// GetRelevantRuleIDs ports Language.getRelevantRules ID lists for SmallLang modules.
// Unknown short codes → nil (no invent).
func (s SmallLang) GetRelevantRuleIDs() []string {
	switch s.ShortCode {
	case "uk":
		return UkrainianRelevantRuleIDs()
	case "gl":
		return GalicianRelevantRuleIDs()
	case "sv":
		return SwedishRelevantRuleIDs()
	// Greek: GreekLang.GetRelevantRuleIDs (greek_relevant_rules.go)
	case "ga":
		return IrishRelevantRuleIDs()
	// Belarusian: BelarusianLang.GetRelevantRuleIDs (belarusian_relevant_rules.go)
	// Breton: BretonLang.GetRelevantRuleIDs (breton_relevant_rules.go)
	case "eo":
		return EsperantoRelevantRuleIDs()
	// Slovak: SlovakLang.GetRelevantRuleIDs (slovak_relevant_rules.go)
	case "da":
		return DanishRelevantRuleIDs()
	// Romanian: RomanianLang.GetRelevantRuleIDs (romanian_relevant_rules.go)
	case "ja":
		return JapaneseRelevantRuleIDs()
	case "zh":
		return ChineseRelevantRuleIDs()
	case "km":
		return KhmerRelevantRuleIDs()
	case "ta":
		return TamilRelevantRuleIDs()
	// Tagalog: TagalogLang.GetRelevantRuleIDs (tagalog_relevant_rules.go)
	case "is":
		return IcelandicRelevantRuleIDs()
	case "ml":
		return MalayalamRelevantRuleIDs()
	case "fa":
		return PersianRelevantRuleIDs()
	case "lt":
		return LithuanianRelevantRuleIDs()
	// CrimeanTatar: CrimeanTatarLang.GetRelevantRuleIDs (crimean_tatar_relevant_rules.go)
	// Asturian: AsturianLang.GetRelevantRuleIDs (asturian_relevant_rules.go)
	// Slovenian: SlovenianLang.GetRelevantRuleIDs (slovenian_relevant_rules.go)
	default:
		return nil
	}
}
