package el

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/redundancies.txt
var redundancyFS embed.FS

// GreekRedundancyReplaceID ports GreekRedundancyRule.EL_REDUNDANCY_REPLACE.
const GreekRedundancyReplaceID = "EL_REDUNDANCY_REPLACE"

// GreekRedundancyFilePath is the official resource path Java loads
// (GreekRedundancyRule.FILE_NAME / getFileNames() → "/el/redundancies.txt").
const GreekRedundancyFilePath = "/el/redundancies.txt"

var (
	redundancyOnce sync.Once
	redundancyBase *rules.AbstractSimpleReplaceRule2
)

// loadRedundancy loads official /el/redundancies.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadRedundancy() *rules.AbstractSimpleReplaceRule2 {
	redundancyOnce.Do(func() {
		f, err := redundancyFS.Open("data/redundancies.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of GreekRedundancyRule + AbstractSimpleReplaceRule2 defaults:
		// - id EL_REDUNDANCY_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator default ", " (Java getSuggestionsSeparator not overridden)
		// - CaseSensitivy.CI (base default)
		// - Locale el via LanguageCode (getLocale → new Locale("el"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.REDUNDANCY, ITSIssueType.Style
		// - example ανεβαίνω πάνω → ανεβαίνω
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          GreekRedundancyReplaceID,
			Description: "Έλεγχος για χρήση πλεονασμού σε μια πρόταση.",
			ShortMsg:    "Πλεονασμός",
			// Java getMessage(): "'$match' είναι πλεονασμός. Γενικά, είναι προτιμότερο το: $suggestions"
			MessageTemplate: "'$match' είναι πλεονασμός. Γενικά, είναι προτιμότερο το: $suggestions",
			// SuggestionsSeparator left empty → Match uses Java default ", "
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "el",
			SubRuleSpecificIDs: false,
			Category:           rules.CatRedundancy.GetCategory(nil),
			IssueType:          rules.ITSStyle,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, GreekRedundancyFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("Μου αρέσει να <marker>ανεβαίνω πάνω</marker> σε δέντρα."),
		//                Example.fixed("Μου αρέσει να <marker>ανεβαίνω</marker> σε δέντρα."));
		base.AddExamplePair(
			rules.Wrong("Μου αρέσει να <marker>ανεβαίνω πάνω</marker> σε δέντρα."),
			rules.Fixed("Μου αρέσει να <marker>ανεβαίνω</marker> σε δέντρα."),
		)
		redundancyBase = base
	})
	return redundancyBase
}

// GreekRedundancyRule ports org.languagetool.rules.el.GreekRedundancyRule
// (extends AbstractSimpleReplaceRule2).
type GreekRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewGreekRedundancyRule ports GreekRedundancyRule(ResourceBundle messages, Language language).
// Official data is go:embed of the same redundancies.txt Java loads from the classpath.
// language is represented via LanguageCode "el" on the base (getLocale → Locale("el")).
func NewGreekRedundancyRule(messages map[string]string) *GreekRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	// Java ctor after super: setCategory(Categories.REDUNDANCY…); setLocQualityIssueType(Style)
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	return &GreekRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports GreekRedundancyRule.getFileNames() → singleton "/el/redundancies.txt".
func (r *GreekRedundancyRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{GreekRedundancyFilePath}
}

// GetShort ports GreekRedundancyRule.getShort().
func (r *GreekRedundancyRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports GreekRedundancyRule.getMessage().
func (r *GreekRedundancyRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports AbstractSimpleReplaceRule2.getSuggestionsSeparator()
// (GreekRedundancyRule does not override) → ", ".
func (r *GreekRedundancyRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports GreekRedundancyRule.getLocale() → Locale("el").
func (r *GreekRedundancyRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from redundancies.txt).
func (r *GreekRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
