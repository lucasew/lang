package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/redundancies.txt
var redundancyFS embed.FS

// EnglishRedundancyReplaceID ports EnglishRedundancyRule.EN_REDUNDANCY_REPLACE.
const EnglishRedundancyReplaceID = "EN_REDUNDANCY_REPLACE"

// EnglishRedundancyFilePath is the official resource path Java loads
// (EnglishRedundancyRule.FILE_NAME / getFileNames() → "/en/redundancies.txt").
const EnglishRedundancyFilePath = "/en/redundancies.txt"

// EnglishRedundancyURL ports EnglishRedundancyRule.getUrl()
// → Tools.getUrl("https://en.wikipedia.org/wiki/Redundancy_(linguistics)").
const EnglishRedundancyURL = "https://en.wikipedia.org/wiki/Redundancy_(linguistics)"

var (
	redundancyOnce sync.Once
	redundancyBase *rules.AbstractSimpleReplaceRule2
)

// loadRedundancy loads official redundancies.txt once (Java AbstractSimpleReplaceRule2 lazyInit).
func loadRedundancy() *rules.AbstractSimpleReplaceRule2 {
	redundancyOnce.Do(func() {
		f, err := redundancyFS.Open("data/redundancies.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of EnglishRedundancyRule + AbstractSimpleReplaceRule2 defaults:
		// - id EN_REDUNDANCY_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " or "
		// - CaseSensitivy.CI (base default)
		// - Locale en via LanguageCode (getLocale → new Locale("en"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.REDUNDANCY, ITSIssueType.Style
		// - setDefaultOff() // see #1997
		// - URL Wikipedia Redundancy (linguistics)
		// - example tuna fish → tuna
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   EnglishRedundancyReplaceID,
			Description:          "1. Redundancy (General)",
			ShortMsg:             "Redundancy",
			// Java getMessage() has no trailing period after $suggestions
			MessageTemplate:      "'$match' is a redundancy. In some cases, it might be preferable to use $suggestions",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "en",
			SubRuleSpecificIDs:   false,
			Category:             rules.CatRedundancy.GetCategory(nil),
			IssueType:            rules.ITSStyle,
			DefaultOff:           true,
			URL:                  EnglishRedundancyURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, EnglishRedundancyFilePath); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("<marker>tuna fish</marker>"),
			rules.Fixed("<marker>tuna</marker>"),
		)
		redundancyBase = base
	})
	return redundancyBase
}

// EnglishRedundancyRule ports org.languagetool.rules.en.EnglishRedundancyRule
// (extends AbstractSimpleReplaceRule2).
type EnglishRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewEnglishRedundancyRule ports EnglishRedundancyRule(ResourceBundle messages).
// Official data is go:embed of the same redundancies.txt Java loads from the classpath.
func NewEnglishRedundancyRule(messages map[string]string) *EnglishRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.DefaultOff = true
	r.URL = EnglishRedundancyURL
	return &EnglishRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports EnglishRedundancyRule.getFileNames() → singleton "/en/redundancies.txt".
func (r *EnglishRedundancyRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{EnglishRedundancyFilePath}
}

// GetShort ports EnglishRedundancyRule.getShort().
func (r *EnglishRedundancyRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports EnglishRedundancyRule.getMessage().
func (r *EnglishRedundancyRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports EnglishRedundancyRule.getSuggestionsSeparator() → " or ".
func (r *EnglishRedundancyRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports EnglishRedundancyRule.getLocale() → Locale("en").
func (r *EnglishRedundancyRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from redundancies.txt).
func (r *EnglishRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
