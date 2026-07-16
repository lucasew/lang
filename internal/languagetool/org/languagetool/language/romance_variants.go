package language

// FrenchVariant ports French locale languages.
type FrenchVariant struct {
	ShortCode     string
	Name          string
	Countries     []string
	SpellerRuleID string
}

func (v FrenchVariant) GetName() string { return v.Name }
func (v FrenchVariant) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

var (
	FrenchFrance = FrenchVariant{ShortCode: "fr", Name: "French", Countries: []string{"FR"}, SpellerRuleID: "MORFOLOGIK_RULE_FR"}
	CanadianFrench = FrenchVariant{ShortCode: "fr-CA", Name: "French (Canada)", Countries: []string{"CA"}, SpellerRuleID: "MORFOLOGIK_RULE_FR"}
	SwissFrench = FrenchVariant{ShortCode: "fr-CH", Name: "French (Switzerland)", Countries: []string{"CH"}, SpellerRuleID: "MORFOLOGIK_RULE_FR"}
	BelgianFrench = FrenchVariant{ShortCode: "fr-BE", Name: "French (Belgium)", Countries: []string{"BE"}, SpellerRuleID: "MORFOLOGIK_RULE_FR"}
)

func AllFrenchVariants() []FrenchVariant {
	return []FrenchVariant{FrenchFrance, CanadianFrench, SwissFrench, BelgianFrench}
}

// SpanishVariant ports Spanish locales.
type SpanishVariant struct {
	ShortCode     string
	Name          string
	Countries     []string
	SpellerRuleID string
	Voseo         bool
}

func (v SpanishVariant) GetName() string { return v.Name }

var (
	SpanishSpain = SpanishVariant{ShortCode: "es", Name: "Spanish", Countries: []string{"ES"}, SpellerRuleID: "MORFOLOGIK_RULE_ES"}
	SpanishVoseo = SpanishVariant{ShortCode: "es-AR", Name: "Spanish (voseo)", Countries: []string{"AR"}, SpellerRuleID: "MORFOLOGIK_RULE_ES", Voseo: true}
)

func AllSpanishVariants() []SpanishVariant {
	return []SpanishVariant{SpanishSpain, SpanishVoseo}
}

// PortugueseVariant ports Portuguese locales.
type PortugueseVariant struct {
	ShortCode     string
	Name          string
	Countries     []string
	SpellerRuleID string
}

func (v PortugueseVariant) GetName() string { return v.Name }

var (
	PortugalPortuguese = PortugueseVariant{ShortCode: "pt-PT", Name: "Portuguese (Portugal)", Countries: []string{"PT"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_PT"}
	BrazilianPortuguese = PortugueseVariant{ShortCode: "pt-BR", Name: "Portuguese (Brazil)", Countries: []string{"BR"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_BR"}
	AngolaPortuguese = PortugueseVariant{ShortCode: "pt-AO", Name: "Portuguese (Angola)", Countries: []string{"AO"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_AO"}
	MozambiquePortuguese = PortugueseVariant{ShortCode: "pt-MZ", Name: "Portuguese (Mozambique)", Countries: []string{"MZ"}, SpellerRuleID: "MORFOLOGIK_RULE_PT_MZ"}
)

func AllPortugueseVariants() []PortugueseVariant {
	return []PortugueseVariant{PortugalPortuguese, BrazilianPortuguese, AngolaPortuguese, MozambiquePortuguese}
}
