package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

// IrishReplaceRuleID ports IrishReplaceRule.getId() → "GA_REPLACE".
const IrishReplaceRuleID = "GA_REPLACE"

// IrishReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/ga/replace.txt")).
const IrishReplaceFilePath = "/ga/replace.txt"

var (
	replaceOnce sync.Once
	replaceMap  map[string][]string
)

// loadReplace ports static wrongWords = loadFromPath("/ga/replace.txt").
// Official resource is go:embed of the same replace.txt Java loads (byte-identical).
func loadReplace() map[string][]string {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		replaceMap = m
	})
	return replaceMap
}

// IrishReplaceRule ports org.languagetool.rules.ga.IrishReplaceRule
// (extends AbstractSimpleReplaceRule, not AbstractSimpleReplaceRule2).
//
// Java field twin summary:
//   - id GA_REPLACE
//   - static wrongWords from /ga/replace.txt
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - isCaseSensitive() false
//   - setCheckLemmas(false)
//   - getLocale() → Locale("GA")
//   - description / short / message hardcoded Irish strings
//   - example: bhúr → bhur
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl
type IrishReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewIrishReplaceRule ports IrishReplaceRule(ResourceBundle messages, Language language).
// language is Irish; getLocale is Locale("GA") — stored as LanguageCode.
func NewIrishReplaceRule(messages map[string]string) *IrishReplaceRule {
	// Field twin of IrishReplaceRule + AbstractSimpleReplaceRule:
	// - id GA_REPLACE
	// - static wrongWords from /ga/replace.txt
	// - isCaseSensitive() false
	// - setCheckLemmas(false)
	// - getLocale → Locale("GA") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - description / short / message hardcoded GA strings
	// - example pair bhúr → bhur
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplace(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            IrishReplaceRuleID,
		LanguageCode:  "GA", // Java: new Locale("GA")
		Description:   "Litriú mícheart, m.sh., \"agsu\" in áit \"agus\"",
		ShortMsg:      "Litriú mícheart",
		Category:      rules.CatTypos.GetCategory(messages),
		IssueType:     rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Litriú mícheart: \"" + String.join(", ", replacements) + "\".";
			// tokenStr is unused (matches Java getMessage signature).
			return "Litriú mícheart: \"" + strings.Join(replacements, ", ") + "\"."
		},
	}
	// Java: bhúr → bhur
	base.AddExamplePair(
		rules.Wrong("Níl beann agam oraibh, ar <marker>bhúr</marker> gcuid cainte."),
		rules.Fixed("Níl beann agam oraibh, ar <marker>bhur</marker> gcuid cainte."),
	)
	return &IrishReplaceRule{AbstractSimpleReplaceRule: base}
}

// joinCommaGA joins suggestion strings with ", " (shared by other GA replace rules).
func joinCommaGA(ss []string) string {
	return strings.Join(ss, ", ")
}

// GetShort ports IrishReplaceRule.getShort() → "Litriú mícheart".
func (r *IrishReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports IrishReplaceRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *IrishReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports IrishReplaceRule.getLocale() → Locale("GA").
func (r *IrishReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports IrishReplaceRule.getWrongWords() → static map.
func (r *IrishReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *IrishReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
