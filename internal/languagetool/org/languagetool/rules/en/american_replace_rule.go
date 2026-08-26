package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/en-US/replace.txt
var usReplaceFS embed.FS

// AmericanSimpleReplaceRuleID ports AmericanReplaceRule.AMERICAN_SIMPLE_REPLACE_RULE.
const AmericanSimpleReplaceRuleID = "EN_US_SIMPLE_REPLACE"

// AmericanReplaceFilePath is the official resource path Java loads
// (AmericanEnglish.getRelevantRules / AmericanReplaceRuleTest → "/en/en-US/replace.txt").
const AmericanReplaceFilePath = "/en/en-US/replace.txt"

var (
	usReplaceOnce sync.Once
	usReplaceBase *rules.AbstractSimpleReplaceRule2
)

// loadUSReplace loads official en-US/replace.txt once (Java lazyInit of AbstractSimpleReplaceRule2).
func loadUSReplace() *rules.AbstractSimpleReplaceRule2 {
	usReplaceOnce.Do(func() {
		f, err := usReplaceFS.Open("data/en-US/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of AmericanReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id AMERICAN_SIMPLE_REPLACE_RULE
		// - description / short / message as Java overrides
		// - CaseSensitivy.CI (base default)
		// - Locale en-US via LanguageCode (getLocale)
		// - useSubRuleSpecificIds()
		// - Categories.STYLE, ITSIssueType.LocaleViolation
		// - example crisps → chips
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                 AmericanSimpleReplaceRuleID,
			Description:        "British words easily confused in American English: $match",
			ShortMsg:           "British word",
			MessageTemplate:    "'$match' is a common British expression. Consider using expressions more common to American English.",
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "en-US",
			SubRuleSpecificIDs: true,
			Category:           rules.CatStyle.GetCategory(nil),
			IssueType:          rules.ITSLocaleViolation,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, AmericanReplaceFilePath); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("Are baked <marker>crisps</marker> healthy?"),
			rules.Fixed("Are baked <marker>chips</marker> healthy?"),
		)
		usReplaceBase = base
	})
	return usReplaceBase
}

// AmericanReplaceRule ports org.languagetool.rules.en.AmericanReplaceRule
// (extends AbstractSimpleReplaceRule2).
type AmericanReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports AmericanReplaceRule.path / getFileNames().
	path string
}

// NewAmericanReplaceRule ports AmericanReplaceRule(ResourceBundle messages, String path).
// Official data is go:embed of the same replace.txt Java loads from the classpath.
func NewAmericanReplaceRule(messages map[string]string, path string) *AmericanReplaceRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("AmericanReplaceRule: path must not be null")
	}
	base := loadUSReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	return &AmericanReplaceRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports AmericanReplaceRule.getFileNames().
func (r *AmericanReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports AmericanReplaceRule.getShort().
func (r *AmericanReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports AmericanReplaceRule.getMessage().
func (r *AmericanReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetLocaleCode ports AmericanReplaceRule.getLocale() → Locale("en-US").
func (r *AmericanReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *AmericanReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
