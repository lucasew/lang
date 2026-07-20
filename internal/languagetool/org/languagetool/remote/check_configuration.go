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
