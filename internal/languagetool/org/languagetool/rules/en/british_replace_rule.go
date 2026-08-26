package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/en-GB/replace.txt
var gbReplaceFS embed.FS

// BritishSimpleReplaceRuleID ports BritishReplaceRule.BRITISH_SIMPLE_REPLACE_RULE.
const BritishSimpleReplaceRuleID = "EN_GB_SIMPLE_REPLACE"

// BritishReplaceFilePath is the official resource path Java loads
// (BritishEnglish.getRelevantRules / BritishReplaceRuleTest → "/en/en-GB/replace.txt").
const BritishReplaceFilePath = "/en/en-GB/replace.txt"

var (
	gbReplaceOnce sync.Once
	gbReplaceBase *rules.AbstractSimpleReplaceRule2
)

// loadGBReplace loads official en-GB/replace.txt once (Java lazyInit of AbstractSimpleReplaceRule2).
func loadGBReplace() *rules.AbstractSimpleReplaceRule2 {
	gbReplaceOnce.Do(func() {
		f, err := gbReplaceFS.Open("data/en-GB/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of BritishReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id BRITISH_SIMPLE_REPLACE_RULE
		// - description / short / message as Java overrides
		// - CaseSensitivy.CI (base default)
		// - Locale en-GB via LanguageCode (getLocale)
		// - useSubRuleSpecificIds()
		// - Categories.STYLE, ITSIssueType.LocaleViolation
		// - example drapes → curtains
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                 BritishSimpleReplaceRuleID,
			Description:        "American words easily confused in British English: $match",
			ShortMsg:           "American word",
			MessageTemplate:    "'$match' is a common American expression. Consider using expressions more common to British English.",
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "en-GB",
			SubRuleSpecificIDs: true,
			Category:           rules.CatStyle.GetCategory(nil),
			IssueType:          rules.ITSLocaleViolation,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, BritishReplaceFilePath); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("We can produce <marker>drapes</marker> of any size or shape from a choice of over 500 different fabrics."),
			rules.Fixed("We can produce <marker>curtains</marker> of any size or shape from a choice of over 500 different fabrics."),
		)
		gbReplaceBase = base
	})
	return gbReplaceBase
}

// BritishReplaceRule ports org.languagetool.rules.en.BritishReplaceRule
// (extends AbstractSimpleReplaceRule2).
type BritishReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports BritishReplaceRule.path / getFileNames().
	path string
}

// NewBritishReplaceRule ports BritishReplaceRule(ResourceBundle messages, String path).
// Official data is go:embed of the same replace.txt Java loads from the classpath.
func NewBritishReplaceRule(messages map[string]string, path string) *BritishReplaceRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("BritishReplaceRule: path must not be null")
	}
	base := loadGBReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	return &BritishReplaceRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports BritishReplaceRule.getFileNames().
func (r *BritishReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports BritishReplaceRule.getShort().
func (r *BritishReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports BritishReplaceRule.getMessage().
func (r *BritishReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetLocaleCode ports BritishReplaceRule.getLocale() → Locale("en-GB").
func (r *BritishReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *BritishReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
