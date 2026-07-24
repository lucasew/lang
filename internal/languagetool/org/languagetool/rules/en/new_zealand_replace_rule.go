package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/en-NZ/replace.txt
var nzReplaceFS embed.FS

// NewZealandSimpleReplaceRuleID ports NewZealandReplaceRule.NEW_ZEALAND_SIMPLE_REPLACE_RULE.
const NewZealandSimpleReplaceRuleID = "EN_NZ_SIMPLE_REPLACE"

// NewZealandReplaceFilePath is the official resource path Java loads
// (NewZealandEnglish.getRelevantRules / NewZealandReplaceRuleTest → "/en/en-NZ/replace.txt").
const NewZealandReplaceFilePath = "/en/en-NZ/replace.txt"

var (
	nzReplaceOnce sync.Once
	nzReplaceBase *rules.AbstractSimpleReplaceRule2
)

// loadNZReplace loads official en-NZ/replace.txt once (Java lazyInit of AbstractSimpleReplaceRule2).
func loadNZReplace() *rules.AbstractSimpleReplaceRule2 {
	nzReplaceOnce.Do(func() {
		f, err := nzReplaceFS.Open("data/en-NZ/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of NewZealandReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id NEW_ZEALAND_SIMPLE_REPLACE_RULE
		// - description / short / message as Java overrides (no "$match" suffix on description)
		// - CaseSensitivy.CI (base default)
		// - Locale en-NZ via LanguageCode (getLocale)
		// - does NOT call useSubRuleSpecificIds() (unlike American/British)
		// - Categories.STYLE, ITSIssueType.LocaleViolation
		// - example sidewalk → footpath
		base := &rules.AbstractSimpleReplaceRule2{
			ID:              NewZealandSimpleReplaceRuleID,
			Description:     "English words easily confused in New Zealand English",
			ShortMsg:        "Not a New Zealand English word",
			MessageTemplate: "'$match' is a non-standard expression. Consider using expressions more common to New Zealand English.",
			CaseSens:        rules.CaseInsensitive,
			LanguageCode:    "en-NZ",
			// SubRuleSpecificIDs remains false — Java does not call useSubRuleSpecificIds().
			Category:  rules.CatStyle.GetCategory(nil),
			IssueType: rules.ITSLocaleViolation,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, NewZealandReplaceFilePath); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("A <marker>sidewalk</marker> is a path along the side of a road."),
			rules.Fixed("A <marker>footpath</marker> is a path along the side of a road."),
		)
		nzReplaceBase = base
	})
	return nzReplaceBase
}

// NewZealandReplaceRule ports org.languagetool.rules.en.NewZealandReplaceRule
// (extends AbstractSimpleReplaceRule2).
type NewZealandReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports NewZealandReplaceRule.path / getFileNames().
	path string
}

// NewNewZealandReplaceRule ports NewZealandReplaceRule(ResourceBundle messages, String path).
// Official data is go:embed of the same replace.txt Java loads from the classpath.
func NewNewZealandReplaceRule(messages map[string]string, path string) *NewZealandReplaceRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("NewZealandReplaceRule: path must not be null")
	}
	base := loadNZReplace()
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	return &NewZealandReplaceRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports NewZealandReplaceRule.getFileNames().
func (r *NewZealandReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports NewZealandReplaceRule.getShort().
func (r *NewZealandReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports NewZealandReplaceRule.getMessage().
func (r *NewZealandReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetLocaleCode ports NewZealandReplaceRule.getLocale() → Locale("en-NZ").
func (r *NewZealandReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *NewZealandReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
