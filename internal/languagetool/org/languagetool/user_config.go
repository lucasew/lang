package languagetool

// UserConfigTokenType ports UserConfig.TokenType.
type UserConfigTokenType string

const (
	TokenInvalid UserConfigTokenType = "INVALID_TOKEN"
	TokenNone    UserConfigTokenType = "NO_TOKEN"
	TokenTest    UserConfigTokenType = "TEST_TOKEN"
	TokenTrial   UserConfigTokenType = "TRIAL_TOKEN"
)

var abTestEnabled bool

func EnableABTests()          { abTestEnabled = true }
func HasABTestsEnabled() bool { return abTestEnabled }

// UserConfig ports a surface subset of org.languagetool.UserConfig.
type UserConfig struct {
	UserSpecificSpellerWords []string
	AcceptedPhrases          map[string]struct{}
	MaxSpellingSuggestions   int
	UserDictName             string
	PremiumUID               *int64
	ConfigurableRuleValues   map[string][]any
	LinguServices            *LinguServices
	FilterDictionaryMatches  bool
	HidePremiumMatches       bool
	TextSessionID            *int64
	ABTest                   []string
	PreferredLanguages       string
	TrustedSource            bool
	OptInThirdPartyAI        bool
	IsPremium                bool
	TokenType                UserConfigTokenType
}

func NewUserConfig() *UserConfig {
	return &UserConfig{
		AcceptedPhrases:        map[string]struct{}{},
		ConfigurableRuleValues: map[string][]any{},
		TokenType:              TokenNone,
	}
}

func (u *UserConfig) GetUserSpecificSpellerWords() []string {
	return u.UserSpecificSpellerWords
}

func (u *UserConfig) AcceptsPhrase(phrase string) bool {
	if u == nil {
		return false
	}
	_, ok := u.AcceptedPhrases[phrase]
	return ok
}

func (u *UserConfig) AddAcceptedPhrase(phrase string) {
	if u.AcceptedPhrases == nil {
		u.AcceptedPhrases = map[string]struct{}{}
	}
	u.AcceptedPhrases[phrase] = struct{}{}
}

func (u *UserConfig) GetConfigValueByID(ruleID string) []any {
	if u == nil {
		return nil
	}
	return u.ConfigurableRuleValues[ruleID]
}

func (u *UserConfig) SetConfigValueByID(ruleID string, values []any) {
	if u.ConfigurableRuleValues == nil {
		u.ConfigurableRuleValues = map[string][]any{}
	}
	u.ConfigurableRuleValues[ruleID] = values
}

func (u *UserConfig) GetLinguServices() *LinguServices {
	if u == nil {
		return nil
	}
	return u.LinguServices
}
