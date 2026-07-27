package km

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via KhmerSimpleReplaceRule:
//
//	getFileNames() → Collections.singletonList("/km/coherency.txt")
//
//go:embed data/coherency.txt
var replaceFS embed.FS

// KhmerSimpleReplaceRuleID ports KhmerSimpleReplaceRule.KHMER_SIMPLE_REPLACE_RULE
// → "KM_SIMPLE_REPLACE".
const KhmerSimpleReplaceRuleID = "KM_SIMPLE_REPLACE"

// KhmerCoherencyFilePath ports KhmerSimpleReplaceRule.FILE_NAME → "/km/coherency.txt".
const KhmerCoherencyFilePath = "/km/coherency.txt"

// KhmerSimpleReplaceFileNames ports KhmerSimpleReplaceRule.getFileNames().
var KhmerSimpleReplaceFileNames = []string{KhmerCoherencyFilePath}

var (
	replaceOnce sync.Once
	replaceBase *rules.AbstractSimpleReplaceRule2
)

// loadReplace loads official /km/coherency.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadReplace() *rules.AbstractSimpleReplaceRule2 {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/coherency.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of KhmerSimpleReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id KM_SIMPLE_REPLACE (KHMER_SIMPLE_REPLACE_RULE)
		// - description / short / message as Java overrides
		// - suggestions separator " or " (Java getSuggestionsSeparator override)
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "km" via LanguageCode (getLocale → new Locale("km"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.MISC (AbstractSimpleReplaceRule2 ctor + KhmerSimpleReplaceRule ctor setCategory)
		// - ITSIssueType.Uncategorized (Rule default; Java does not setLocQualityIssueType)
		// - no addExamplePair in Java
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   KhmerSimpleReplaceRuleID,
			Description:          "Words or groups of words that are incorrect or obsolete",
			ShortMsg:             "Consider following the spelling of Chuon Nath",
			// Java getMessage(): " Consider following the spelling of Chuon Nath " (leading/trailing spaces)
			MessageTemplate:      " Consider following the spelling of Chuon Nath ",
			SuggestionsSeparator: " or ", // Java getSuggestionsSeparator override
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "km", // Java: new Locale("km")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatMisc.GetCategory(nil),
			// IssueType left empty → GetLocQualityIssueType returns ITSUncategorized
		}
		if err := base.LoadSimpleReplaceRule2Data(f, KhmerCoherencyFilePath); err != nil {
			panic(err)
		}
		replaceBase = base
	})
	return replaceBase
}

// KhmerSimpleReplaceRule ports org.languagetool.rules.km.KhmerSimpleReplaceRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /km/coherency.txt.
type KhmerSimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewKhmerSimpleReplaceRule ports KhmerSimpleReplaceRule(ResourceBundle messages).
// Java: super(messages, new Khmer()); setCategory(Categories.MISC.getCategory(messages)).
// Official data is go:embed of the same coherency.txt Java loads from the classpath.
// language is represented via LanguageCode "km" on the base (getLocale → Locale("km")).
func NewKhmerSimpleReplaceRule(messages map[string]string) *KhmerSimpleReplaceRule {
	base := loadReplace()
	r := *base
	r.Messages = messages
	// Java AbstractSimpleReplaceRule2 ctor + KhmerSimpleReplaceRule ctor:
	// setCategory(Categories.MISC.getCategory(messages))
	r.Category = rules.CatMisc.GetCategory(messages)
	return &KhmerSimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports KhmerSimpleReplaceRule.getFileNames()
// → Collections.singletonList("/km/coherency.txt").
func (r *KhmerSimpleReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{KhmerCoherencyFilePath}
}

// GetShort ports KhmerSimpleReplaceRule.getShort().
func (r *KhmerSimpleReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports KhmerSimpleReplaceRule.getMessage().
func (r *KhmerSimpleReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports KhmerSimpleReplaceRule.getSuggestionsSeparator() → " or ".
func (r *KhmerSimpleReplaceRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Khmer always sets " or " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports KhmerSimpleReplaceRule.getLocale() → Locale("km").
func (r *KhmerSimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from coherency.txt).
func (r *KhmerSimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
