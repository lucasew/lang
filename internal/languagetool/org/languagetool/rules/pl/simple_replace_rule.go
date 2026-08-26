package pl

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via SimpleReplaceRule:
//
//	static wrongWords = loadFromPath("/pl/replace.txt")
//
//go:embed data/replace.txt
var replaceFS embed.FS

// PolishSimpleReplaceRuleID ports SimpleReplaceRule.POLISH_SIMPLE_REPLACE_RULE.
const PolishSimpleReplaceRuleID = "PL_SIMPLE_REPLACE"

// PolishReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/pl/replace.txt")).
const PolishReplaceFilePath = "/pl/replace.txt"

var (
	replaceOnce  sync.Once
	replaceWords map[string][]string
)

// loadReplaceWords ports static wrongWords = loadFromPath("/pl/replace.txt").
// Official resource is go:embed of the same replace.txt Java loads (byte-identical).
func loadReplaceWords() map[string][]string {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			// Java getFromRulesDirAsStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		replaceWords = m
	})
	return replaceWords
}

// SimpleReplaceRule ports org.languagetool.rules.pl.SimpleReplaceRule
// (extends AbstractSimpleReplaceRule).
//
// Java field twin summary:
//   - id PL_SIMPLE_REPLACE (POLISH_SIMPLE_REPLACE_RULE)
//   - static wrongWords from /pl/replace.txt
//   - setCategory(new Category(new CategoryId("PRAWDOPODOBNE_LITEROWKI"), "Prawdopodobne literówki"))
//   - setLocQualityIssueType(ITSIssueType.Misspelling)
//   - setCheckLemmas(false)
//   - isCaseSensitive() false
//   - getLocale() → Locale("pl")
//   - description / short / message hardcoded Polish strings
//   - example pair: sei → się
//   - no setIgnoreTaggedWords, no useSubRuleSpecificIds, no setUrl, no setDefaultOff
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewSimpleReplaceRule ports SimpleReplaceRule(ResourceBundle messages, Language language).
// language is represented via LanguageCode "pl" (getLocale → Locale("pl")).
func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplaceWords(),
		CaseSensitive: false, // Java isCaseSensitive() false
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            PolishSimpleReplaceRuleID,
		LanguageCode:  "pl", // Java: new Locale("pl")
		Description:   "Typowe literówki i niepoprawne wyrazy (domowi, sie, niewiadomo, duh, cie…)",
		ShortMsg:      "Literówka",
		// Java: setCategory(new Category(new CategoryId("PRAWDOPODOBNE_LITEROWKI"), "Prawdopodobne literówki"))
		Category: rules.NewCategory(
			rules.NewCategoryId("PRAWDOPODOBNE_LITEROWKI"),
			"Prawdopodobne literówki",
		),
		// Java: setLocQualityIssueType(ITSIssueType.Misspelling)
		IssueType: rules.ITSMisspelling,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java: "Wyraz „" + tokenStr + "” to najczęściej literówka; poprawnie pisze się: "
			//       + String.join(", ", replacements) + "."
			return "Wyraz „" + tokenStr + "” to najczęściej literówka; poprawnie pisze się: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	// Java:
	// addExamplePair(Example.wrong("Uspokój <marker>sei</marker>."),
	//                Example.fixed("Uspokój <marker>się</marker>."));
	base.AddExamplePair(
		rules.Wrong("Uspokój <marker>sei</marker>."),
		rules.Fixed("Uspokój <marker>się</marker>."),
	)
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports SimpleReplaceRule.getShort() → "Literówka".
func (r *SimpleReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleReplaceRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real suggestions.
func (r *SimpleReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports SimpleReplaceRule.getLocale() → Locale("pl").
func (r *SimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports SimpleReplaceRule.getWrongWords() → static map.
func (r *SimpleReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// IsDefaultOff ports Rule.isDefaultOff — Java never calls setDefaultOff.
func (r *SimpleReplaceRule) IsDefaultOff() bool {
	return false
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
