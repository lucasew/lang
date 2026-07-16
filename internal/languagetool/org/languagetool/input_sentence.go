package languagetool

// InputSentence ports org.languagetool.InputSentence — cache key for check results.
// Language fields use short codes (Language interface deferred).
type InputSentence struct {
	Analyzed           *AnalyzedSentence
	LanguageCode       string
	MotherTongueCode   string
	DisabledRules      map[string]struct{}
	DisabledCategories map[string]struct{}
	EnabledRules       map[string]struct{}
	EnabledCategories  map[string]struct{}
	UserConfig         *UserConfig
	AltLanguageCodes   []string
	Mode               string // JLanguageTool.Mode name
	Level              Level
	TextSessionID      *int64
	ToneTags           map[ToneTag]struct{}
}

func NewInputSentence(
	analyzed *AnalyzedSentence,
	languageCode, motherTongueCode string,
	disabledRules, disabledCategories, enabledRules, enabledCategories map[string]struct{},
	userConfig *UserConfig,
	altLanguages []string,
	mode string,
	level Level,
	textSessionID *int64,
	toneTags map[ToneTag]struct{},
) InputSentence {
	if languageCode == "" {
		panic("language required")
	}
	if mode == "" {
		panic("mode required")
	}
	tt := toneTags
	if tt == nil {
		tt = map[ToneTag]struct{}{}
	}
	return InputSentence{
		Analyzed:           analyzed,
		LanguageCode:       languageCode,
		MotherTongueCode:   motherTongueCode,
		DisabledRules:      disabledRules,
		DisabledCategories: disabledCategories,
		EnabledRules:       enabledRules,
		EnabledCategories:  enabledCategories,
		UserConfig:         userConfig,
		AltLanguageCodes:   append([]string(nil), altLanguages...),
		Mode:               mode,
		Level:              level,
		TextSessionID:      textSessionID,
		ToneTags:           tt,
	}
}

func (s InputSentence) GetAnalyzedSentence() *AnalyzedSentence { return s.Analyzed }

func (s InputSentence) String() string {
	if s.Analyzed == nil {
		return ""
	}
	return s.Analyzed.GetText()
}

func (s InputSentence) Equal(o InputSentence) bool {
	if s.LanguageCode != o.LanguageCode || s.MotherTongueCode != o.MotherTongueCode {
		return false
	}
	if s.Mode != o.Mode || s.Level != o.Level {
		return false
	}
	st, ot := "", ""
	if s.Analyzed != nil {
		st = s.Analyzed.GetText()
	}
	if o.Analyzed != nil {
		ot = o.Analyzed.GetText()
	}
	if st != ot {
		return false
	}
	if (s.TextSessionID == nil) != (o.TextSessionID == nil) {
		return false
	}
	if s.TextSessionID != nil && *s.TextSessionID != *o.TextSessionID {
		return false
	}
	return stringSetEqual(s.DisabledRules, o.DisabledRules) &&
		stringSetEqual(s.DisabledCategories, o.DisabledCategories) &&
		stringSetEqual(s.EnabledRules, o.EnabledRules) &&
		stringSetEqual(s.EnabledCategories, o.EnabledCategories) &&
		stringSliceEqual(s.AltLanguageCodes, o.AltLanguageCodes) &&
		toneSetEqual(s.ToneTags, o.ToneTags)
}

func stringSetEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func toneSetEqual(a, b map[ToneTag]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}
