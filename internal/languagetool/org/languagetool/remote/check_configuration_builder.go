package remote

// CheckConfigurationBuilder ports org.languagetool.remote.CheckConfigurationBuilder.
type CheckConfigurationBuilder struct {
	langCode             string
	motherTongueLangCode string
	autoDetectLanguage   bool
	enabledOnly          bool
	enabledRuleIDs       []string
	disabledRuleIDs      []string
	mode                 string
	level                string
	ruleValues           []string
	textSessionID        string
	username             string
	apiKey               string
}

// NewCheckConfigurationBuilder builds a config for a fixed language code.
func NewCheckConfigurationBuilder(langCode string) *CheckConfigurationBuilder {
	if langCode == "" {
		panic("langCode must not be empty")
	}
	return &CheckConfigurationBuilder{langCode: langCode}
}

// NewAutoDetectCheckConfigurationBuilder activates server-side language guessing.
func NewAutoDetectCheckConfigurationBuilder() *CheckConfigurationBuilder {
	return &CheckConfigurationBuilder{autoDetectLanguage: true}
}

func (b *CheckConfigurationBuilder) Build() *CheckConfiguration {
	if b == nil {
		panic("nil builder")
	}
	if b.enabledOnly && len(b.enabledRuleIDs) == 0 {
		panic("You cannot use 'enabledOnly' when you haven't set rule ids to be enabled")
	}
	if b.langCode == "" && !b.autoDetectLanguage {
		panic("No language was set but language guessing was not activated either")
	}
	if b.langCode != "" && b.autoDetectLanguage {
		panic("Language was set but language guessing was also activated")
	}
	return &CheckConfiguration{
		LangCode:             b.langCode,
		MotherTongueLangCode: b.motherTongueLangCode,
		GuessLanguage:        b.autoDetectLanguage,
		EnabledRuleIDs:       append([]string(nil), b.enabledRuleIDs...),
		EnabledOnly:          b.enabledOnly,
		DisabledRuleIDs:      append([]string(nil), b.disabledRuleIDs...),
		Mode:                 b.mode,
		Level:                b.level,
		RuleValues:           append([]string(nil), b.ruleValues...),
		TextSessionID:        b.textSessionID,
		Username:             b.username,
		APIKey:               b.apiKey,
	}
}

func (b *CheckConfigurationBuilder) SetMotherTongueLangCode(code string) *CheckConfigurationBuilder {
	b.motherTongueLangCode = code
	return b
}

func (b *CheckConfigurationBuilder) EnabledRuleIDs(ids ...string) *CheckConfigurationBuilder {
	b.enabledRuleIDs = append([]string(nil), ids...)
	return b
}

func (b *CheckConfigurationBuilder) DisabledRuleIDs(ids ...string) *CheckConfigurationBuilder {
	b.disabledRuleIDs = append([]string(nil), ids...)
	return b
}

func (b *CheckConfigurationBuilder) EnabledOnly() *CheckConfigurationBuilder {
	b.enabledOnly = true
	return b
}

func (b *CheckConfigurationBuilder) Mode(mode string) *CheckConfigurationBuilder {
	b.mode = mode
	return b
}

func (b *CheckConfigurationBuilder) Level(level string) *CheckConfigurationBuilder {
	b.level = level
	return b
}

func (b *CheckConfigurationBuilder) RuleValues(values ...string) *CheckConfigurationBuilder {
	b.ruleValues = append([]string(nil), values...)
	return b
}

func (b *CheckConfigurationBuilder) TextSessionID(id string) *CheckConfigurationBuilder {
	b.textSessionID = id
	return b
}

func (b *CheckConfigurationBuilder) Username(u string) *CheckConfigurationBuilder {
	b.username = u
	return b
}

func (b *CheckConfigurationBuilder) APIKey(k string) *CheckConfigurationBuilder {
	b.apiKey = k
	return b
}
