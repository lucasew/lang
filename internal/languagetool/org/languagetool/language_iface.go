package languagetool

// Language is the pure-Go surface for org.languagetool.Language (minimal).
// Concrete languages implement this without requiring the full Java hierarchy.
type Language interface {
	GetName() string
	GetShortCode() string
	GetShortCodeWithCountryAndVariant() string
	GetCountries() []string
	// CreateDefaultTagger may return nil if untagged.
	CreateDefaultTagger() Tagger
}

// LanguageDefaults provides default hooks shared by many languages.
type LanguageDefaults struct {
	Name      string
	ShortCode string
	Countries []string
	Tagger    Tagger
	// Typography when advanced typography is enabled for the language.
	Typography TypographyConfig
	// MaintainedState ports getMaintainedState.
	MaintainedState LanguageMaintainedState
	// ConsistencyRulePrefix ports getConsistencyRulePrefix.
	ConsistencyRulePrefix string
}

func (l LanguageDefaults) GetName() string {
	if l.Name == "" {
		return l.ShortCode
	}
	return l.Name
}
func (l LanguageDefaults) GetShortCode() string { return l.ShortCode }
func (l LanguageDefaults) GetShortCodeWithCountryAndVariant() string {
	return l.ShortCode
}
func (l LanguageDefaults) GetCountries() []string      { return l.Countries }
func (l LanguageDefaults) CreateDefaultTagger() Tagger { return l.Tagger }

func (l LanguageDefaults) GetMaintainedState() LanguageMaintainedState {
	if l.MaintainedState == "" {
		return LookingForNewMaintainer
	}
	return l.MaintainedState
}

func (l LanguageDefaults) GetConsistencyRulePrefix() string {
	if l.ConsistencyRulePrefix == "" {
		return "PREFIXFORCONSISTENCYRULES_"
	}
	return l.ConsistencyRulePrefix
}

func (l LanguageDefaults) IsAdvancedTypographyEnabled() bool {
	return l.Typography.Enabled
}

func (l LanguageDefaults) ToAdvancedTypography(input string) string {
	return ToAdvancedTypography(input, l.Typography)
}

// HasMinMatchesRules ports Language.hasMinMatchesRules (default false).
func (l LanguageDefaults) HasMinMatchesRules() bool { return false }
