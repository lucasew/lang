package remote

// CheckConfiguration ports org.languagetool.remote.CheckConfiguration.
type CheckConfiguration struct {
	LangCode             string
	MotherTongueLangCode string
	GuessLanguage        bool
	EnabledRuleIDs       []string
	EnabledOnly          bool
	DisabledRuleIDs      []string
	Mode                 string
	Level                string
	RuleValues           []string
	TextSessionID        string
	Username             string
	APIKey               string
}

// NewCheckConfiguration ports the package-private Java constructor (via tests).
// Panics on invalid lang/guess pairs (IllegalArgumentException) and treats nil
// rule-id/value slices as empty (Objects.requireNonNull → empty lists allowed via builder).
func NewCheckConfiguration(
	langCode, motherTongueLangCode string,
	guessLanguage bool,
	enabledRuleIDs []string,
	enabledOnly bool,
	disabledRuleIDs []string,
	mode, level string,
	ruleValues []string,
	textSessionID, username, apiKey string,
) *CheckConfiguration {
	if langCode == "" && !guessLanguage {
		panic("No language was set but language guessing was not activated either")
	}
	if langCode != "" && guessLanguage {
		panic("Language was set but language guessing was also activated")
	}
	if enabledRuleIDs == nil {
		enabledRuleIDs = []string{}
	}
	if disabledRuleIDs == nil {
		disabledRuleIDs = []string{}
	}
	if ruleValues == nil {
		ruleValues = []string{}
	}
	return &CheckConfiguration{
		LangCode:             langCode,
		MotherTongueLangCode: motherTongueLangCode,
		GuessLanguage:        guessLanguage,
		EnabledRuleIDs:       append([]string(nil), enabledRuleIDs...),
		EnabledOnly:          enabledOnly,
		DisabledRuleIDs:      append([]string(nil), disabledRuleIDs...),
		Mode:                 mode,
		Level:                level,
		RuleValues:           append([]string(nil), ruleValues...),
		TextSessionID:        textSessionID,
		Username:             username,
		APIKey:               apiKey,
	}
}

func (c *CheckConfiguration) GetLangCode() (string, bool) {
	if c == nil || c.LangCode == "" {
		return "", false
	}
	return c.LangCode, true
}

func (c *CheckConfiguration) GetMotherTongueLangCode() string {
	if c == nil {
		return ""
	}
	return c.MotherTongueLangCode
}

func (c *CheckConfiguration) IsGuessLanguage() bool {
	return c != nil && c.GuessLanguage
}

func (c *CheckConfiguration) GetEnabledRuleIDs() []string {
	if c == nil {
		return nil
	}
	return append([]string(nil), c.EnabledRuleIDs...)
}

func (c *CheckConfiguration) IsEnabledOnly() bool {
	return c != nil && c.EnabledOnly
}

func (c *CheckConfiguration) GetDisabledRuleIDs() []string {
	if c == nil {
		return nil
	}
	return append([]string(nil), c.DisabledRuleIDs...)
}

func (c *CheckConfiguration) GetMode() string {
	if c == nil {
		return ""
	}
	return c.Mode
}

func (c *CheckConfiguration) GetLevel() string {
	if c == nil {
		return ""
	}
	return c.Level
}

func (c *CheckConfiguration) GetRuleValues() []string {
	if c == nil {
		return nil
	}
	return append([]string(nil), c.RuleValues...)
}

// GetTextSessionID ports CheckConfiguration.getTextSessionID (@Nullable).
func (c *CheckConfiguration) GetTextSessionID() (string, bool) {
	if c == nil || c.TextSessionID == "" {
		return "", false
	}
	return c.TextSessionID, true
}

// GetUsername ports CheckConfiguration.getUsername (@Nullable).
func (c *CheckConfiguration) GetUsername() (string, bool) {
	if c == nil || c.Username == "" {
		return "", false
	}
	return c.Username, true
}

// GetAPIKey ports CheckConfiguration.getAPIKey (@Nullable).
func (c *CheckConfiguration) GetAPIKey() (string, bool) {
	if c == nil || c.APIKey == "" {
		return "", false
	}
	return c.APIKey, true
}
