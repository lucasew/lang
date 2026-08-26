package languagetool

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

// UserConfigTokenType ports UserConfig.TokenType.
type UserConfigTokenType string

const (
	TokenInvalid UserConfigTokenType = "INVALID_TOKEN"
	TokenNone    UserConfigTokenType = "NO_TOKEN"
	TokenTest    UserConfigTokenType = "TEST_TOKEN"
	TokenTrial   UserConfigTokenType = "TRIAL_TOKEN"
)

var abTestEnabled bool

// EnableABTests ports UserConfig.enableABTests.
func EnableABTests() { abTestEnabled = true }

// HasABTestsEnabled ports UserConfig.hasABTestsEnabled.
func HasABTestsEnabled() bool { return abTestEnabled }

// UserConfig ports org.languagetool.UserConfig.
type UserConfig struct {
	UserSpecificSpellerWords []string
	AcceptedPhrases          map[string]struct{}
	// UserSpecificRules holds Rule twins as any (rules package cycle avoidance).
	UserSpecificRules       []any
	MaxSpellingSuggestions  int
	UserDictCacheSize       *int64
	UserDictName            string
	PremiumUID              *int64
	ConfigurableRuleValues  map[string][]any
	LinguServices           *LinguServices
	FilterDictionaryMatches bool
	HidePremiumMatches      bool
	TextSessionID           *int64
	ABTest                  []string
	// preferredLanguages is the cleaned/sorted comma-joined main-lang string (Java field).
	PreferredLanguages string
	TrustedSource      bool
	OptInThirdPartyAI  bool
	IsPremium          bool
	TokenType          UserConfigTokenType
	SuggestionsEnabled bool
}

// NewUserConfig ports UserConfig() empty constructor.
func NewUserConfig() *UserConfig {
	return NewUserConfigWithWords(nil, nil)
}

// NewUserConfigWithWords ports UserConfig(List words, Map ruleValues).
func NewUserConfigWithWords(words []string, ruleValues map[string][]any) *UserConfig {
	if words == nil {
		words = []string{}
	}
	if ruleValues == nil {
		ruleValues = map[string][]any{}
	}
	u := &UserConfig{
		UserSpecificSpellerWords: append([]string(nil), words...),
		UserSpecificRules:        []any{},
		ConfigurableRuleValues:   map[string][]any{},
		UserDictName:             "default",
		TokenType:                TokenNone,
		SuggestionsEnabled:       true,
	}
	for k, v := range ruleValues {
		u.ConfigurableRuleValues[k] = v
	}
	u.AcceptedPhrases = buildAcceptedPhrases(u.UserSpecificSpellerWords)
	return u
}

func buildAcceptedPhrases(words []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, w := range words {
		if strings.Contains(w, " ") {
			out[w] = struct{}{}
		}
	}
	return out
}

// removeAllButMainLanguagesAndSort ports private helper.
// Main languages: de, en, es, fr, nl, pt. Returns joined string if >=2, else "".
func removeAllButMainLanguagesAndSort(preferred []string) string {
	if preferred == nil {
		return ""
	}
	main := map[string]struct{}{"de": {}, "en": {}, "es": {}, "fr": {}, "nl": {}, "pt": {}}
	var clean []string
	for _, language := range preferred {
		if _, ok := main[language]; ok {
			clean = append(clean, language)
		}
	}
	sort.Strings(clean)
	if len(clean) >= 2 {
		return strings.Join(clean, ",")
	}
	return ""
}

// SetPreferredLanguagesList applies removeAllButMainLanguagesAndSort (constructor field).
func (u *UserConfig) SetPreferredLanguagesList(langs []string) {
	if u == nil {
		return
	}
	u.PreferredLanguages = removeAllButMainLanguagesAndSort(langs)
}

// GetAcceptedWords ports getAcceptedWords.
func (u *UserConfig) GetAcceptedWords() []string {
	if u == nil {
		return []string{}
	}
	return u.UserSpecificSpellerWords
}

// GetUserSpecificSpellerWords is a Go alias for GetAcceptedWords (existing callers).
func (u *UserConfig) GetUserSpecificSpellerWords() []string { return u.GetAcceptedWords() }

// GetAcceptedPhrases ports getAcceptedPhrases.
func (u *UserConfig) GetAcceptedPhrases() map[string]struct{} {
	if u == nil {
		return map[string]struct{}{}
	}
	if u.AcceptedPhrases == nil {
		return map[string]struct{}{}
	}
	return u.AcceptedPhrases
}

// AcceptsPhrase is a convenience helper (not a Java method).
func (u *UserConfig) AcceptsPhrase(phrase string) bool {
	if u == nil {
		return false
	}
	_, ok := u.AcceptedPhrases[phrase]
	return ok
}

// AddAcceptedPhrase is a Go helper for tests; rebuilds phrases from speller words if needed.
func (u *UserConfig) AddAcceptedPhrase(phrase string) {
	if u.AcceptedPhrases == nil {
		u.AcceptedPhrases = map[string]struct{}{}
	}
	u.AcceptedPhrases[phrase] = struct{}{}
}

// GetRules ports getRules (Rule list as any).
func (u *UserConfig) GetRules() []any {
	if u == nil {
		return []any{}
	}
	return u.UserSpecificRules
}

func (u *UserConfig) GetMaxSpellingSuggestions() int {
	if u == nil {
		return 0
	}
	return u.MaxSpellingSuggestions
}

// GetConfigValues ports getConfigValues.
func (u *UserConfig) GetConfigValues() map[string][]any {
	if u == nil {
		return map[string][]any{}
	}
	return u.ConfigurableRuleValues
}

// InsertConfigValues ports insertConfigValues.
func (u *UserConfig) InsertConfigValues(ruleValues map[string][]any) {
	if u == nil {
		return
	}
	if u.ConfigurableRuleValues == nil {
		u.ConfigurableRuleValues = map[string][]any{}
	}
	for k, v := range ruleValues {
		u.ConfigurableRuleValues[k] = v
	}
}

func (u *UserConfig) GetConfigValueByID(ruleID string) []any {
	if u == nil || u.ConfigurableRuleValues == nil {
		return nil
	}
	v, ok := u.ConfigurableRuleValues[ruleID]
	if !ok {
		return nil
	}
	return v
}

func (u *UserConfig) SetConfigValueByID(ruleID string, values []any) {
	if u.ConfigurableRuleValues == nil {
		u.ConfigurableRuleValues = map[string][]any{}
	}
	u.ConfigurableRuleValues[ruleID] = values
}

func (u *UserConfig) HasLinguServices() bool {
	return u != nil && u.LinguServices != nil
}

func (u *UserConfig) GetLinguServices() *LinguServices {
	if u == nil {
		return nil
	}
	return u.LinguServices
}

func (u *UserConfig) GetUserDictCacheSize() *int64 {
	if u == nil {
		return nil
	}
	return u.UserDictCacheSize
}

func (u *UserConfig) GetUserDictName() string {
	if u == nil {
		return "default"
	}
	return u.UserDictName
}

func (u *UserConfig) GetPremiumUid() *int64 {
	if u == nil {
		return nil
	}
	return u.PremiumUID
}

func (u *UserConfig) IsSuggestionsEnabled() bool {
	if u == nil {
		return true
	}
	return u.SuggestionsEnabled
}

func (u *UserConfig) GetTextSessionId() *int64 {
	if u == nil {
		return nil
	}
	return u.TextSessionID
}

func (u *UserConfig) GetAbTest() []string {
	if u == nil {
		return nil
	}
	return u.ABTest
}

func (u *UserConfig) GetHidePremiumMatches() bool {
	if u == nil {
		return false
	}
	return u.HidePremiumMatches
}

// GetPreferredLanguages ports getPreferredLanguages — split of cleaned string.
// Note: Java Arrays.asList("".split(",")) yields [""] one empty element.
func (u *UserConfig) GetPreferredLanguages() []string {
	if u == nil {
		return []string{""}
	}
	return strings.Split(u.PreferredLanguages, ",")
}

// Equal ports equals (EqualsBuilder fields).
func (u *UserConfig) Equal(o *UserConfig) bool {
	if u == o {
		return true
	}
	if u == nil || o == nil {
		return false
	}
	if !mapsAnyEqual(u.ConfigurableRuleValues, o.ConfigurableRuleValues) {
		return false
	}
	if ruleIDHashSum(u.UserSpecificRules) != ruleIDHashSum(o.UserSpecificRules) {
		return false
	}
	if !int64PtrEqual(u.PremiumUID, o.PremiumUID) {
		return false
	}
	if u.UserDictName != o.UserDictName {
		return false
	}
	if !stringSliceEqual(u.UserSpecificSpellerWords, o.UserSpecificSpellerWords) {
		return false
	}
	if u.FilterDictionaryMatches != o.FilterDictionaryMatches {
		return false
	}
	if !stringSliceEqual(u.ABTest, o.ABTest) {
		return false
	}
	if u.HidePremiumMatches != o.HidePremiumMatches {
		return false
	}
	if u.PreferredLanguages != o.PreferredLanguages {
		return false
	}
	if u.TrustedSource != o.TrustedSource {
		return false
	}
	if u.OptInThirdPartyAI != o.OptInThirdPartyAI {
		return false
	}
	if u.IsPremium != o.IsPremium {
		return false
	}
	if u.TokenType != o.TokenType {
		return false
	}
	if u.SuggestionsEnabled != o.SuggestionsEnabled {
		return false
	}
	return true
}

// Hash ports hashCode (HashCodeBuilder fields; omits speller words).
func (u *UserConfig) Hash() uint64 {
	if u == nil {
		return 0
	}
	h := fnv.New64a()
	writeI64(h, int64(u.MaxSpellingSuggestions))
	writeI64(h, ruleIDHashSum(u.UserSpecificRules))
	if u.PremiumUID != nil {
		writeI64(h, *u.PremiumUID)
	}
	_, _ = h.Write([]byte(u.UserDictName))
	if u.UserDictCacheSize != nil {
		writeI64(h, *u.UserDictCacheSize)
	}
	// configurableRuleValues — order keys
	keys := make([]string, 0, len(u.ConfigurableRuleValues))
	for k := range u.ConfigurableRuleValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte(fmt.Sprint(u.ConfigurableRuleValues[k])))
	}
	for _, a := range u.ABTest {
		_, _ = h.Write([]byte(a))
	}
	if u.FilterDictionaryMatches {
		_, _ = h.Write([]byte{1})
	}
	if u.HidePremiumMatches {
		_, _ = h.Write([]byte{1})
	}
	_, _ = h.Write([]byte(u.PreferredLanguages))
	if u.TrustedSource {
		_, _ = h.Write([]byte{1})
	}
	if u.OptInThirdPartyAI {
		_, _ = h.Write([]byte{1})
	}
	if u.IsPremium {
		_, _ = h.Write([]byte{1})
	}
	_, _ = h.Write([]byte(u.TokenType))
	if u.SuggestionsEnabled {
		_, _ = h.Write([]byte{1})
	}
	return h.Sum64()
}

func (u *UserConfig) String() string {
	if u == nil {
		return "UserConfig{}"
	}
	return fmt.Sprintf("UserConfig{dictionarySize=%d, maxSpellingSuggestions=%d, userDictName='%s', configurableRuleValues=%v, linguServices=%v, filterDictionaryMatches=%v, textSessionId=%v, hidePremiumMatches=%v, abTest='%v', optInThirdPartyAI=%v, suggestionsEnabled=%v}",
		len(u.UserSpecificSpellerWords), u.MaxSpellingSuggestions, u.UserDictName, u.ConfigurableRuleValues, u.LinguServices, u.FilterDictionaryMatches, u.TextSessionID, u.HidePremiumMatches, u.ABTest, u.OptInThirdPartyAI, u.SuggestionsEnabled)
}

// ruleWithID is optional interface for UserSpecificRules equals/hash (Java Rule.getId()).
type ruleWithID interface {
	GetID() string
}

func ruleIDHashSum(rules []any) int64 {
	var sum int64
	for _, r := range rules {
		if rw, ok := r.(ruleWithID); ok {
			// Java: k.getId().hashCode() — approximate with FNV32
			h := fnv.New32a()
			_, _ = h.Write([]byte(rw.GetID()))
			sum += int64(int32(h.Sum32())) // Java int hash may be negative
		}
	}
	return sum
}

func int64PtrEqual(a, b *int64) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func mapsAnyEqual(a, b map[string][]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		ov, ok := b[k]
		if !ok || len(v) != len(ov) {
			return false
		}
		for i := range v {
			if fmt.Sprint(v[i]) != fmt.Sprint(ov[i]) {
				return false
			}
		}
	}
	return true
}

type hasher interface {
	Write(p []byte) (n int, err error)
}

func writeI64(h hasher, n int64) {
	var buf [8]byte
	for i := 0; i < 8; i++ {
		buf[i] = byte(n >> (8 * i))
	}
	_, _ = h.Write(buf[:])
}
