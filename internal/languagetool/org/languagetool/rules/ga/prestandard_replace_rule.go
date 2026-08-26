package ga

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace-prestandard.txt
var prestandardFS embed.FS

// PrestandardReplaceRuleID ports PrestandardReplaceRule.getId() → "GA_PRESTANDARD_REPLACE".
const PrestandardReplaceRuleID = "GA_PRESTANDARD_REPLACE"

// PrestandardReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/ga/replace-prestandard.txt")).
const PrestandardReplaceFilePath = "/ga/replace-prestandard.txt"

var (
	prestandardOnce sync.Once
	prestandardMap  map[string][]string
)

// loadPrestandard ports static wrongWords = loadFromPath("/ga/replace-prestandard.txt").
// Official resource is go:embed of the same replace-prestandard.txt Java loads (byte-identical).
func loadPrestandard() map[string][]string {
	prestandardOnce.Do(func() {
		f, err := prestandardFS.Open("data/replace-prestandard.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		prestandardMap = m
	})
	return prestandardMap
}

// PrestandardReplaceRule ports org.languagetool.rules.ga.PrestandardReplaceRule
// (extends AbstractSimpleReplaceRule).
//
// Java field twin summary:
//   - id GA_PRESTANDARD_REPLACE
//   - static wrongWords from /ga/replace-prestandard.txt
//   - Categories.TYPOS (overrides ASR ctor MISC)
//   - ITSIssueType.Misspelling
//   - isCaseSensitive() false
//   - setCheckLemmas(false)
//   - getLocale() → Locale("GA")
//   - description / short / message hardcoded Irish strings
//   - message has no space after colon: Litriú réamhchaighdeánach:"…"
//   - example: baoghal → baol
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl
type PrestandardReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewPrestandardReplaceRule ports PrestandardReplaceRule(ResourceBundle messages, Language language).
// language is Irish; getLocale is Locale("GA") — stored as LanguageCode.
func NewPrestandardReplaceRule(messages map[string]string) *PrestandardReplaceRule {
	// Field twin of PrestandardReplaceRule + AbstractSimpleReplaceRule:
	// - id GA_PRESTANDARD_REPLACE
	// - static wrongWords from /ga/replace-prestandard.txt
	// - isCaseSensitive() false
	// - setCheckLemmas(false)
	// - getLocale → Locale("GA") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - description / short / message hardcoded GA strings
	// - example pair baoghal → baol
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadPrestandard(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            PrestandardReplaceRuleID,
		LanguageCode:  "GA", // Java: new Locale("GA")
		Description:   "Litriú réamhchaighdeánach, m.sh., \"baoghal\" in áit \"baol\"",
		ShortMsg:      "Litriú réamhchaighdeánach",
		Category:      rules.CatTypos.GetCategory(messages),
		IssueType:     rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage:
			// return "Litriú réamhchaighdeánach:\""
			//     + String.join(", ", replacements) + "\".";
			// No space after colon — match Java exactly.
			// tokenStr is unused (matches Java getMessage signature).
			return "Litriú réamhchaighdeánach:\"" + strings.Join(replacements, ", ") + "\"."
		},
	}
	// Java: baoghal → baol
	base.AddExamplePair(
		rules.Wrong("“Ní <marker>baoghal</marker> daoibh,” ar sise."),
		rules.Fixed("“Ní <marker>baol</marker> daoibh,” ar sise."),
	)
	return &PrestandardReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports PrestandardReplaceRule.getShort() → "Litriú réamhchaighdeánach".
func (r *PrestandardReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PrestandardReplaceRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *PrestandardReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports PrestandardReplaceRule.getLocale() → Locale("GA").
func (r *PrestandardReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports PrestandardReplaceRule.getWrongWords() → static map.
func (r *PrestandardReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *PrestandardReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
