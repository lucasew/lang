package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace-fgb-eq.txt
var fgbEqFS embed.FS

// IrishFGBEqReplaceRuleID ports IrishFGBEqReplaceRule.getId() → "GA_FGB_EQ_REPLACE".
const IrishFGBEqReplaceRuleID = "GA_FGB_EQ_REPLACE"

// IrishFGBEqReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/ga/replace-fgb-eq.txt")).
const IrishFGBEqReplaceFilePath = "/ga/replace-fgb-eq.txt"

var (
	fgbEqOnce sync.Once
	fgbEqMap  map[string][]string
)

// loadFGBEq ports static wrongWords = loadFromPath("/ga/replace-fgb-eq.txt").
// Official resource is go:embed of the same replace-fgb-eq.txt Java loads (byte-identical).
func loadFGBEq() map[string][]string {
	fgbEqOnce.Do(func() {
		f, err := fgbEqFS.Open("data/replace-fgb-eq.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		fgbEqMap = m
	})
	return fgbEqMap
}

// IrishFGBEqReplaceRule ports org.languagetool.rules.ga.IrishFGBEqReplaceRule
// (extends AbstractSimpleReplaceRule).
//
// Java field twin summary:
//   - id GA_FGB_EQ_REPLACE
//   - static wrongWords from /ga/replace-fgb-eq.txt
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - isCaseSensitive() false
//   - setCheckLemmas(false)
//   - getLocale() → Locale("GA")
//   - description / short / message hardcoded Irish strings
//   - example: urlamh → ullamh
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl
type IrishFGBEqReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewIrishFGBEqReplaceRule ports IrishFGBEqReplaceRule(ResourceBundle messages, Language language).
// language is Irish; getLocale is Locale("GA") — stored as LanguageCode.
func NewIrishFGBEqReplaceRule(messages map[string]string) *IrishFGBEqReplaceRule {
	// Field twin of IrishFGBEqReplaceRule + AbstractSimpleReplaceRule:
	// - id GA_FGB_EQ_REPLACE
	// - static wrongWords from /ga/replace-fgb-eq.txt
	// - isCaseSensitive() false
	// - setCheckLemmas(false)
	// - getLocale → Locale("GA") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - description / short / message hardcoded GA strings
	// - example pair urlamh → ullamh
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadFGBEq(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            IrishFGBEqReplaceRuleID,
		LanguageCode:  "GA", // Java: new Locale("GA")
		Description:   "Ceannfhocal FGB neamhchoitianta, m.sh., \"urlamh\" in áit \"ullamh\"",
		ShortMsg:      "Neamhchoitianta",
		Category:      rules.CatTypos.GetCategory(messages),
		IssueType:     rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Focal ceart ach tá \""
			//     + String.join(", ", replacements) + "\" níos coitianta.";
			// tokenStr is unused (matches Java getMessage signature).
			return "Focal ceart ach tá \"" + strings.Join(replacements, ", ") + "\" níos coitianta."
		},
	}
	// Java: urlamh → ullamh
	base.AddExamplePair(
		rules.Wrong("An bhfuil tú <marker>urlamh</marker>?"),
		rules.Fixed("An bhfuil tú <marker>ullamh</marker>?"),
	)
	return &IrishFGBEqReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports IrishFGBEqReplaceRule.getShort() → "Neamhchoitianta".
func (r *IrishFGBEqReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports IrishFGBEqReplaceRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *IrishFGBEqReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports IrishFGBEqReplaceRule.getLocale() → Locale("GA").
func (r *IrishFGBEqReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports IrishFGBEqReplaceRule.getWrongWords() → static map.
func (r *IrishFGBEqReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *IrishFGBEqReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
