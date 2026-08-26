package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/spaces.txt
var spacesFS embed.FS

// SpacesRuleID ports SpacesRule.getId() → "GA_SPASANNA".
const SpacesRuleID = "GA_SPASANNA"

// SpacesFilePath is the official classpath resource Java loads
// (static loadFromPath("/ga/spaces.txt")).
const SpacesFilePath = "/ga/spaces.txt"

var (
	spacesOnce sync.Once
	spacesMap  map[string][]string
)

// loadSpaces ports static wrongWords = loadFromPath("/ga/spaces.txt").
// Official resource is go:embed of the same spaces.txt Java loads (byte-identical).
func loadSpaces() map[string][]string {
	spacesOnce.Do(func() {
		f, err := spacesFS.Open("data/spaces.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		spacesMap = m
	})
	return spacesMap
}

// SpacesRule ports org.languagetool.rules.ga.SpacesRule
// (extends AbstractSimpleReplaceRule; missing spaces between words).
//
// Java field twin summary:
//   - id GA_SPASANNA
//   - static wrongWords from /ga/spaces.txt
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - setIgnoreTaggedWords() true
//   - setCheckLemmas(false)
//   - isCaseSensitive() false
//   - getLocale() → Locale("GA")
//   - description / short / message hardcoded Irish strings
//   - no addExamplePair, no setUrl, no useSubRuleSpecificIds
type SpacesRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSpacesRule ports SpacesRule(ResourceBundle messages, Language language).
// language is Irish; getLocale is Locale("GA") — stored as LanguageCode.
func NewSpacesRule(messages map[string]string) *SpacesRule {
	// Field twin of SpacesRule + AbstractSimpleReplaceRule:
	// - id GA_SPASANNA
	// - static wrongWords from /ga/spaces.txt
	// - isCaseSensitive() false
	// - setCheckLemmas(false)
	// - setIgnoreTaggedWords() true
	// - getLocale → Locale("GA") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - description / short / message hardcoded GA strings
	// - no AddExamplePair (Java has none), no setUrl
	base := &rules.AbstractSimpleReplaceRule{
		Messages:          messages,
		WrongWords:        loadSpaces(),
		CaseSensitive:     false, // Java isCaseSensitive() false
		CheckLemmas:       false, // Java setCheckLemmas(false)
		IgnoreTaggedWords: true,  // Java setIgnoreTaggedWords()
		ID:                SpacesRuleID,
		LanguageCode:      "GA", // Java: new Locale("GA")
		Description:       "Spás ar iarraidh",
		ShortMsg:          "Spás ar iarraidh",
		Category:          rules.CatTypos.GetCategory(messages),
		IssueType:         rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Spás ar iarraidh: \""
			//     + String.join(", ", replacements) + "\".";
			// tokenStr is unused (matches Java getMessage signature).
			return "Spás ar iarraidh: \"" + strings.Join(replacements, ", ") + "\"."
		},
	}
	// Java has no addExamplePair — do not invent Example pairs.
	return &SpacesRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SpacesRule.getShort() → "Spás ar iarraidh".
func (r *SpacesRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SpacesRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *SpacesRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports SpacesRule.getLocale() → Locale("GA").
func (r *SpacesRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports SpacesRule.getWrongWords() → static map.
func (r *SpacesRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SpacesRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
