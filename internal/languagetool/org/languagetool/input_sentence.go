package languagetool

import "hash/fnv"

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

// NewInputSentence ports the full constructor (with textSessionID + toneTags).
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
	// Java: Objects.requireNonNull(lang); Objects.requireNonNull(mode); Objects.requireNonNull(level)
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

// NewInputSentenceFromUserConfig ports ctor that takes textSessionId from userConfig.
func NewInputSentenceFromUserConfig(
	analyzed *AnalyzedSentence,
	languageCode, motherTongueCode string,
	disabledRules, disabledCategories, enabledRules, enabledCategories map[string]struct{},
	userConfig *UserConfig,
	altLanguages []string,
	mode string,
	level Level,
	toneTags map[ToneTag]struct{},
) InputSentence {
	var sid *int64
	if userConfig != nil {
		sid = userConfig.GetTextSessionId()
	}
	return NewInputSentence(analyzed, languageCode, motherTongueCode,
		disabledRules, disabledCategories, enabledRules, enabledCategories,
		userConfig, altLanguages, mode, level, sid, toneTags)
}

func (s InputSentence) GetAnalyzedSentence() *AnalyzedSentence { return s.Analyzed }

func (s InputSentence) String() string {
	if s.Analyzed == nil {
		return ""
	}
	return s.Analyzed.GetText()
}

// Equal ports equals — includes userConfig (Java Objects.equals(userConfig, other.userConfig)).
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
	// UserConfig equals by value
	if s.UserConfig == nil && o.UserConfig == nil {
		// ok
	} else if s.UserConfig == nil || o.UserConfig == nil {
		return false
	} else if !s.UserConfig.Equal(o.UserConfig) {
		return false
	}
	// nil vs empty alt languages: Java Objects.equals treats empty list vs null as unequal
	// but tests expect empty slice ~ nil for Go convenience when both empty length 0.
	// Match Java list equality: nil and empty both length 0 and iterate equal.
	return stringSetEqual(s.DisabledRules, o.DisabledRules) &&
		stringSetEqual(s.DisabledCategories, o.DisabledCategories) &&
		stringSetEqual(s.EnabledRules, o.EnabledRules) &&
		stringSetEqual(s.EnabledCategories, o.EnabledCategories) &&
		stringSliceEqual(s.AltLanguageCodes, o.AltLanguageCodes) &&
		toneSetEqual(s.ToneTags, o.ToneTags)
}

// Hash ports hashCode.
func (s InputSentence) Hash() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.String()))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(s.LanguageCode))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(s.MotherTongueCode))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(s.Mode))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(string(s.Level)))
	if s.UserConfig != nil {
		writeI64(h, int64(s.UserConfig.Hash()))
	}
	if s.TextSessionID != nil {
		writeI64(h, *s.TextSessionID)
	}
	return h.Sum64()
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
