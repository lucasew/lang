package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official homophones list Java loads via ArabicHomophonesRule:
//
//	getFileNames() → Collections.singletonList("ar/homophones.txt")
//
//go:embed data/homophones.txt
var homophonesFS embed.FS

// ArabicHomophonesRuleID ports ArabicHomophonesRule.AR_HOMOPHONES_REPLACE → "AR_HOMOPHONES_REPLACE".
const ArabicHomophonesRuleID = "AR_HOMOPHONES_REPLACE"

// ArabicHomophonesFilePath ports ArabicHomophonesRule.RESOURCE_FILENAME → "ar/homophones.txt"
// (no leading slash — matches Java resource path exactly).
const ArabicHomophonesFilePath = "ar/homophones.txt"

// ArabicHomophonesFileNames ports ArabicHomophonesRule.getFileNames().
var ArabicHomophonesFileNames = []string{ArabicHomophonesFilePath}

var (
	homophonesOnce sync.Once
	homophonesBase *rules.AbstractSimpleReplaceRule2
)

// loadHomophones loads official ar/homophones.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadHomophones() *rules.AbstractSimpleReplaceRule2 {
	homophonesOnce.Do(func() {
		f, err := homophonesFS.Open("data/homophones.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicHomophonesRule + AbstractSimpleReplaceRule2 defaults:
		// - id AR_HOMOPHONES_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " أو  " (leading space, two trailing spaces)
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "ar" via LanguageCode (getLocale → new Locale("ar"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.CONFUSED_WORDS (ArabicHomophonesRule ctor after super)
		// - ITSIssueType.Uncategorized (Rule default; Java does not setLocQualityIssueType)
		// - example ضن → ظن
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          ArabicHomophonesRuleID,
			Description: "كلمات متشابهة لفظا للتوضيح، يرجى التحقق منها مثل تشابه الظاء والضاد.",
			// Java getShort()
			ShortMsg: "كلمات متشابهة لفظا يرجى التحقق منها",
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
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicHomophonesFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>ضن</marker>"),
		//   Example.fixed("<marker>ظن</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>ضن</marker>"),
			rules.Fixed("<marker>ظن</marker>"),
		)
		homophonesBase = base
	})
	return homophonesBase
}

// ArabicHomophonesRule ports org.languagetool.rules.ar.ArabicHomophonesRule
// (extends AbstractSimpleReplaceRule2). Loads fixed ar/homophones.txt.
type ArabicHomophonesRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewArabicHomophonesRule ports ArabicHomophonesRule(ResourceBundle messages).
// Java: super(messages, new Arabic()); setCategory(Categories.CONFUSED_WORDS.getCategory(messages)).
// Official data is go:embed of the same homophones.txt Java loads from the classpath.
// language is represented via LanguageCode "ar" on the base (getLocale → Locale("ar")).
func NewArabicHomophonesRule(messages map[string]string) *ArabicHomophonesRule {
	base := loadHomophones()
	r := *base
	r.Messages = messages
	// Java ArabicHomophonesRule ctor after super:
	// super.setCategory(Categories.CONFUSED_WORDS.getCategory(messages));
	r.Category = rules.CatConfusedWords.GetCategory(messages)
	return &ArabicHomophonesRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports ArabicHomophonesRule.getFileNames()
// → Collections.singletonList("ar/homophones.txt").
func (r *ArabicHomophonesRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicHomophonesFilePath}
}

// GetShort ports ArabicHomophonesRule.getShort().
func (r *ArabicHomophonesRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicHomophonesRule.getMessage() → "قل $suggestions".
func (r *ArabicHomophonesRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports ArabicHomophonesRule.getSuggestionsSeparator()
// → " أو  " (leading space, two trailing spaces).
func (r *ArabicHomophonesRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Arabic always sets " أو  " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports ArabicHomophonesRule.getLocale() → Locale("ar").
func (r *ArabicHomophonesRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from homophones.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Khmer/Greek ASR2 leaves — no rebind).
func (r *ArabicHomophonesRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
