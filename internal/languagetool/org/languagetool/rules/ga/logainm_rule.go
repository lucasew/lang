package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/placenames.txt
var placenamesFS embed.FS

// LogainmRuleID ports LogainmRule.getId() → "GA_LOGAINM".
const LogainmRuleID = "GA_LOGAINM"

// LogainmFilePath is the official classpath resource Java loads
// (static loadFromPath("/ga/placenames.txt")).
const LogainmFilePath = "/ga/placenames.txt"

// LogainmURL ports LogainmRule.getUrl() → Tools.getUrl("https://www.logainm.ie/ga/").
const LogainmURL = "https://www.logainm.ie/ga/"

var (
	placenamesOnce sync.Once
	placenamesMap  map[string][]string
)

// loadPlacenames ports static wrongWords = loadFromPath("/ga/placenames.txt").
// Official resource is go:embed of the same placenames.txt Java loads (byte-identical).
func loadPlacenames() map[string][]string {
	placenamesOnce.Do(func() {
		f, err := placenamesFS.Open("data/placenames.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		placenamesMap = m
	})
	return placenamesMap
}

// LogainmRule ports org.languagetool.rules.ga.LogainmRule
// (extends AbstractSimpleReplaceRule; English placenames → Irish).
//
// Java field twin summary:
//   - id GA_LOGAINM
//   - static wrongWords from /ga/placenames.txt
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - setIgnoreTaggedWords() true
//   - setCheckLemmas(false)
//   - isCaseSensitive() false
//   - getLocale() → Locale("GA")
//   - description / short / message hardcoded Irish strings
//   - URL https://www.logainm.ie/ga/
//   - no addExamplePair, no useSubRuleSpecificIds
type LogainmRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewLogainmRule ports LogainmRule(ResourceBundle messages, Language language).
// language is Irish; getLocale is Locale("GA") — stored as LanguageCode.
func NewLogainmRule(messages map[string]string) *LogainmRule {
	// Field twin of LogainmRule + AbstractSimpleReplaceRule:
	// - id GA_LOGAINM
	// - static wrongWords from /ga/placenames.txt
	// - isCaseSensitive() false
	// - setCheckLemmas(false)
	// - setIgnoreTaggedWords() true
	// - getLocale → Locale("GA") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - URL https://www.logainm.ie/ga/
	// - description / short / message hardcoded GA strings
	// - no AddExamplePair (Java has none)
	base := &rules.AbstractSimpleReplaceRule{
		Messages:          messages,
		WrongWords:        loadPlacenames(),
		CaseSensitive:     false, // Java isCaseSensitive() false
		CheckLemmas:       false, // Java setCheckLemmas(false)
		IgnoreTaggedWords: true,  // Java setIgnoreTaggedWords()
		ID:                LogainmRuleID,
		LanguageCode:      "GA", // Java: new Locale("GA")
		Description:       "Logainm Béarla, m.sh., 'Dublin' in áit 'Baile Átha Cliath'.",
		ShortMsg:          "Logainm Béarla",
		Category:          rules.CatTypos.GetCategory(messages),
		IssueType:         rules.ITSMisspelling,
		URL:               LogainmURL,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Is logainm Béarla é \"" + tokenStr + "\" ar \""
			//     + String.join(", ", replacements) + "\".";
			// Both tokenStr and replacements are used (unlike IrishReplace/FGBEq).
			return "Is logainm Béarla é \"" + tokenStr + "\" ar \"" +
				strings.Join(replacements, ", ") + "\"."
		},
	}
	// Java has no addExamplePair — do not invent Example pairs.
	return &LogainmRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports LogainmRule.getShort() → "Logainm Béarla".
func (r *LogainmRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports LogainmRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real token and suggestions.
func (r *LogainmRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports LogainmRule.getLocale() → Locale("GA").
func (r *LogainmRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports LogainmRule.getWrongWords() → static map.
func (r *LogainmRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *LogainmRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
