package language

// Italian language metadata.
var Italian = struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}{ShortCode: "it", Name: "Italian", Countries: []string{"IT"}, SpellerRuleID: "MORFOLOGIK_RULE_IT"}

// Dutch variants.
type DutchVariant struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

func (v DutchVariant) GetName() string { return v.Name }

var (
	DutchNetherlands = DutchVariant{ShortCode: "nl", Name: "Dutch", Countries: []string{"NL"}, SpellerRuleID: "MORFOLOGIK_RULE_NL"}
	BelgianDutch     = DutchVariant{ShortCode: "nl-BE", Name: "Dutch (Belgium)", Countries: []string{"BE"}, SpellerRuleID: "MORFOLOGIK_RULE_NL"}
)

func AllDutchVariants() []DutchVariant {
	return []DutchVariant{DutchNetherlands, BelgianDutch}
}

// Polish metadata.
var Polish = struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}{ShortCode: "pl", Name: "Polish", Countries: []string{"PL"}, SpellerRuleID: "MORFOLOGIK_RULE_PL"}

// Catalan variants.
type CatalanVariant struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
	Valencian                      bool
}

func (v CatalanVariant) GetName() string { return v.Name }

var (
	Catalan          = CatalanVariant{ShortCode: "ca", Name: "Catalan", Countries: []string{"ES"}, SpellerRuleID: "MORFOLOGIK_RULE_CA"}
	ValencianCatalan = CatalanVariant{ShortCode: "ca-ES-valencia", Name: "Catalan (Valencian)", Countries: []string{"ES"}, SpellerRuleID: "MORFOLOGIK_RULE_CA", Valencian: true}
)

func AllCatalanVariants() []CatalanVariant {
	return []CatalanVariant{Catalan, ValencianCatalan}
}

// Russian metadata.
var Russian = struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}{ShortCode: "ru", Name: "Russian", Countries: []string{"RU"}, SpellerRuleID: "MORFOLOGIK_RULE_RU"}
