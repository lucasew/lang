package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via ArabicSimpleReplaceRule:
//
//	getFileNames() → Collections.singletonList("ar/replaces.txt")
//
//go:embed data/replaces.txt
var replacesFS embed.FS

// ArabicSimpleReplaceRuleID ports ArabicSimpleReplaceRule.RULE_ID → "AR_SIMPLE_REPLACE".
const ArabicSimpleReplaceRuleID = "AR_SIMPLE_REPLACE"

// ArabicSimpleReplaceFilePath ports ArabicSimpleReplaceRule.RESOURCE_FILENAME → "ar/replaces.txt"
// (no leading slash — matches Java resource path exactly).
const ArabicSimpleReplaceFilePath = "ar/replaces.txt"

// ArabicSimpleReplaceFileNames ports ArabicSimpleReplaceRule.getFileNames().
var ArabicSimpleReplaceFileNames = []string{ArabicSimpleReplaceFilePath}

var (
	replacesOnce sync.Once
	replacesBase *rules.AbstractSimpleReplaceRule2
)

// loadReplaces loads official ar/replaces.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadReplaces() *rules.AbstractSimpleReplaceRule2 {
	replacesOnce.Do(func() {
		f, err := replacesFS.Open("data/replaces.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicSimpleReplaceRule + AbstractSimpleReplaceRule2 defaults:
		// - id AR_SIMPLE_REPLACE (RULE_ID)
		// - description / short / message as Java overrides
		// - suggestions separator " أو  " (leading space, two trailing spaces)
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "ar" via LanguageCode (getLocale → new Locale("ar"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.CONFUSED_WORDS (ArabicSimpleReplaceRule ctor after super)
		// - ITSIssueType.Uncategorized (Rule default; Java does not setLocQualityIssueType)
		// - isDictionaryBasedSpellingRule() false (Java override; same as Rule default)
		// - example الى → إلى
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          ArabicSimpleReplaceRuleID,
			Description: "قاعدة تطابق الكلمات التي يجب تجنبها وتقترح تصويبا لها",
			// Java getShort(): double space between أن and يقال
			ShortMsg: "خطأ، يفضل أن  يقال:",
			// Java getMessage(): "قل $suggestions"
			MessageTemplate: "قل $suggestions",
			// Java getSuggestionsSeparator(): " أو  " (space + أو + two spaces)
			SuggestionsSeparator: " أو  ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar", // Java: new Locale("ar")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatConfusedWords.GetCategory(nil),
			// IssueType left empty → GetLocQualityIssueType returns ITSUncategorized
		}
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicSimpleReplaceFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>الى</marker>"),
		//   Example.fixed("<marker>إلى</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>الى</marker>"),
			rules.Fixed("<marker>إلى</marker>"),
		)
		replacesBase = base
	})
	return replacesBase
}

// ArabicSimpleReplaceRule ports org.languagetool.rules.ar.ArabicSimpleReplaceRule
// (extends AbstractSimpleReplaceRule2). Loads fixed ar/replaces.txt.
type ArabicSimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewArabicSimpleReplaceRule ports ArabicSimpleReplaceRule(ResourceBundle messages).
// Java: super(messages, new Arabic()); setCategory(Categories.CONFUSED_WORDS.getCategory(messages)).
// Official data is go:embed of the same replaces.txt Java loads from the classpath.
// language is represented via LanguageCode "ar" on the base (getLocale → Locale("ar")).
func NewArabicSimpleReplaceRule(messages map[string]string) *ArabicSimpleReplaceRule {
	base := loadReplaces()
	r := *base
	r.Messages = messages
	// Java ArabicSimpleReplaceRule ctor after super:
	// super.setCategory(Categories.CONFUSED_WORDS.getCategory(messages));
	r.Category = rules.CatConfusedWords.GetCategory(messages)
	return &ArabicSimpleReplaceRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports ArabicSimpleReplaceRule.getFileNames()
// → Collections.singletonList("ar/replaces.txt").
func (r *ArabicSimpleReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicSimpleReplaceFilePath}
}

// GetShort ports ArabicSimpleReplaceRule.getShort().
func (r *ArabicSimpleReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicSimpleReplaceRule.getMessage() → "قل $suggestions".
func (r *ArabicSimpleReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports ArabicSimpleReplaceRule.getSuggestionsSeparator()
// → " أو  " (leading space, two trailing spaces).
func (r *ArabicSimpleReplaceRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Arabic always sets " أو  " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports ArabicSimpleReplaceRule.getLocale() → Locale("ar").
func (r *ArabicSimpleReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// IsDictionaryBasedSpellingRule ports ArabicSimpleReplaceRule.isDictionaryBasedSpellingRule()
// → false (explicit Java override; same as Rule default).
func (r *ArabicSimpleReplaceRule) IsDictionaryBasedSpellingRule() bool {
	return false
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replaces.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Khmer/Greek ASR2 leaves — no rebind).
func (r *ArabicSimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
