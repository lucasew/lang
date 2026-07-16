package language

// SmallLang is metadata for languages ported with tagger/speller surfaces only.
type SmallLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

func (s SmallLang) GetName() string { return s.Name }
func (s SmallLang) GetShortCode() string { return s.ShortCode }

var (
	Slovak     = SmallLang{"sk", "Slovak", "MORFOLOGIK_RULE_SK_SK", []string{"SK"}}
	Danish     = SmallLang{"da", "Danish", "MORFOLOGIK_RULE_DA_DK", []string{"DK"}}
	Swedish    = SmallLang{"sv", "Swedish", "MORFOLOGIK_RULE_SV_SE", []string{"SE"}}
	Romanian   = SmallLang{"ro", "Romanian", "MORFOLOGIK_RULE_RO_RO", []string{"RO"}}
	Greek      = SmallLang{"el", "Greek", "MORFOLOGIK_RULE_EL_GR", []string{"GR"}}
	Galician   = SmallLang{"gl", "Galician", "MORFOLOGIK_RULE_GL_ES", []string{"ES"}}
	Japanese   = SmallLang{"ja", "Japanese", "MORFOLOGIK_RULE_JA", []string{"JP"}}
	Chinese    = SmallLang{"zh", "Chinese", "MORFOLOGIK_RULE_ZH", []string{"CN"}}
	Persian    = SmallLang{"fa", "Persian", "MORFOLOGIK_RULE_FA", []string{"IR"}}
	Esperanto  = SmallLang{"eo", "Esperanto", "MORFOLOGIK_RULE_EO", nil}
	Irish      = SmallLang{"ga", "Irish", "MORFOLOGIK_RULE_GA", []string{"IE"}}
	Ukrainian  = SmallLang{"uk", "Ukrainian", "MORFOLOGIK_RULE_UK_UA", []string{"UA"}}
)

func AllSmallLangs() []SmallLang {
	return []SmallLang{
		Slovak, Danish, Swedish, Romanian, Greek, Galician,
		Japanese, Chinese, Persian, Esperanto, Irish, Ukrainian,
	}
}
