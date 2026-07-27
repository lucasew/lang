package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/english-homophones.txt
var englishHomophonesFS embed.FS

// EnglishHomophoneRuleID ports EnglishHomophoneRule.getId() → "GA_ENGLISH_HOMOPHONE".
const EnglishHomophoneRuleID = "GA_ENGLISH_HOMOPHONE"

// EnglishHomophoneFilePath is the official classpath resource Java loads
// (static loadFromPath("/ga/english-homophones.txt")).
const EnglishHomophoneFilePath = "/ga/english-homophones.txt"

var (
	englishHomophonesOnce sync.Once
	englishHomophonesMap  map[string][]string
)

// loadEnglishHomophones ports static wrongWords = loadFromPath("/ga/english-homophones.txt").
// Official resource is go:embed of the same english-homophones.txt Java loads (byte-identical).
func loadEnglishHomophones() map[string][]string {
	englishHomophonesOnce.Do(func() {
		f, err := englishHomophonesFS.Open("data/english-homophones.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		englishHomophonesMap = m
	})
	return englishHomophonesMap
}

// EnglishHomophoneRule ports org.languagetool.rules.ga.EnglishHomophoneRule
// (extends AbstractSimpleReplaceRule; English words that sound like Irish → Irish spelling).
//
// Java field twin summary:
//   - id GA_ENGLISH_HOMOPHONE
//   - static wrongWords from /ga/english-homophones.txt
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - isCaseSensitive() false
//   - setCheckLemmas(false)
//   - getLocale() → Locale("GA")
//   - description / short / message hardcoded Irish strings
//   - example: sushi → súisí
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl
type EnglishHomophoneRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewEnglishHomophoneRule ports EnglishHomophoneRule(ResourceBundle messages, Language language).
// language is Irish; getLocale is Locale("GA") — stored as LanguageCode.
func NewEnglishHomophoneRule(messages map[string]string) *EnglishHomophoneRule {
	// Field twin of EnglishHomophoneRule + AbstractSimpleReplaceRule:
	// - id GA_ENGLISH_HOMOPHONE
	// - static wrongWords from /ga/english-homophones.txt
	// - isCaseSensitive() false
	// - setCheckLemmas(false)
	// - getLocale → Locale("GA") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - description / short / message hardcoded GA strings
	// - example pair sushi → súisí
	// - no setIgnoreTaggedWords, no setUrl, no useSubRuleSpecificIds
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadEnglishHomophones(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            EnglishHomophoneRuleID,
		LanguageCode:  "GA", // Java: new Locale("GA")
		Description:   "Is homafóin iad na focail, m.sh., \"well\" agus \"bhuel\"",
		ShortMsg:      "Homafón Béarla.",
		Category:      rules.CatTypos.GetCategory(messages),
		IssueType:     rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Is é " + String.join(", ", replacements)
			//     + " an litriú Gaelach ar \"" + tokenStr + "\"";
			// Both tokenStr and replacements are used; no trailing period.
			return "Is é " + strings.Join(replacements, ", ") +
				" an litriú Gaelach ar \"" + tokenStr + "\""
		},
	}
	// Java: sushi → súisí
	base.AddExamplePair(
		rules.Wrong("An bhialann <marker>sushi</marker> sin ba chúis leis."),
		rules.Fixed("An bhialann <marker>súisí</marker> sin ba chúis leis."),
	)
	return &EnglishHomophoneRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports EnglishHomophoneRule.getShort() → "Homafón Béarla.".
func (r *EnglishHomophoneRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports EnglishHomophoneRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real token and suggestions.
func (r *EnglishHomophoneRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports EnglishHomophoneRule.getLocale() → Locale("GA").
func (r *EnglishHomophoneRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports EnglishHomophoneRule.getWrongWords() → static map.
func (r *EnglishHomophoneRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *EnglishHomophoneRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
