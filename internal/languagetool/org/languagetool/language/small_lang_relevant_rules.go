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
	case "el":
		return GreekRelevantRuleIDs()
	case "ga":
		return IrishRelevantRuleIDs()
	case "be":
		return BelarusianRelevantRuleIDs()
	case "br":
		return BretonRelevantRuleIDs()
	case "eo":
		return EsperantoRelevantRuleIDs()
	case "sk":
		return SlovakRelevantRuleIDs()
	case "da":
		return DanishRelevantRuleIDs()
	case "ro":
		return RomanianRelevantRuleIDs()
	case "ja":
		return JapaneseRelevantRuleIDs()
	case "zh":
		return ChineseRelevantRuleIDs()
	case "km":
		return KhmerRelevantRuleIDs()
	case "ta":
		return TamilRelevantRuleIDs()
	case "tl":
		return TagalogRelevantRuleIDs()
	case "is":
		return IcelandicRelevantRuleIDs()
	case "ml":
		return MalayalamRelevantRuleIDs()
	case "fa":
		return PersianRelevantRuleIDs()
	case "lt":
		return LithuanianRelevantRuleIDs()
	case "crh":
		return CrimeanTatarRelevantRuleIDs()
	case "ast":
		return AsturianRelevantRuleIDs()
	case "sl":
		return SlovenianRelevantRuleIDs()
	default:
		return nil
	}
}
