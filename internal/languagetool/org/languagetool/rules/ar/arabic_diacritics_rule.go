package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official diacritics list Java loads via ArabicDiacriticsRule:
//
//	getFileNames() → Collections.singletonList("/ar/diacritics.txt")
//
//go:embed data/diacritics.txt
var diacriticsFS embed.FS

// ArabicDiacriticsRuleID ports ArabicDiacriticsRule.AR_DIACRITICS_REPLACE → "AR_DIACRITICS_REPLACE".
const ArabicDiacriticsRuleID = "AR_DIACRITICS_REPLACE"

// ArabicDiacriticsFilePath ports ArabicDiacriticsRule.FILE_NAME → "/ar/diacritics.txt"
// (leading slash — matches Java resource path exactly).
const ArabicDiacriticsFilePath = "/ar/diacritics.txt"

// ArabicDiacriticsFileNames ports ArabicDiacriticsRule.getFileNames().
var ArabicDiacriticsFileNames = []string{ArabicDiacriticsFilePath}

var (
	diacriticsOnce sync.Once
	diacriticsBase *rules.AbstractSimpleReplaceRule2
)

// loadDiacritics loads official /ar/diacritics.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadDiacritics() *rules.AbstractSimpleReplaceRule2 {
	diacriticsOnce.Do(func() {
		f, err := diacriticsFS.Open("data/diacritics.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicDiacriticsRule + AbstractSimpleReplaceRule2 defaults:
		// - id AR_DIACRITICS_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " أو " (space + أو + one trailing space)
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "ar" via LanguageCode (getLocale → new Locale("ar"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.STYLE (ArabicDiacriticsRule ctor after super)
		// - ITSIssueType.Style (setLocQualityIssueType)
		// - example تجربة → تجرِِبة (Java source has DOUBLE kasra in example — bug-for-bug;
		//   data file maps تجربة=تجرِبة with single kasra)
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          ArabicDiacriticsRuleID,
			Description: "كلمات مشكولة للتوضيح",
			// Java getShort()
			ShortMsg: "كلمات يستحسن أن تشكّل لتصحيح نطقها",
			// Java getMessage(): "'$match' كلمة يشيع نطقها نطقا خاطئا لذا نقترح تشكيلها كالآتي: $suggestions"
			MessageTemplate: "'$match' كلمة يشيع نطقها نطقا خاطئا لذا نقترح تشكيلها كالآتي: $suggestions",
			// Java getSuggestionsSeparator(): " أو " (space + أو + one trailing space)
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar", // Java: new Locale("ar")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatStyle.GetCategory(nil),
			IssueType:            rules.ITSStyle,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicDiacriticsFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>تجربة</marker>"),
		//                Example.fixed("<marker>تجرِِبة</marker>"));
		// Note: double kasra in fixed example is an upstream quirk (data has single).
		base.AddExamplePair(
			rules.Wrong("<marker>تجربة</marker>"),
			rules.Fixed("<marker>تجرِِبة</marker>"),
		)
		diacriticsBase = base
	})
	return diacriticsBase
}

// ArabicDiacriticsRule ports org.languagetool.rules.ar.ArabicDiacriticsRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /ar/diacritics.txt.
type ArabicDiacriticsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewArabicDiacriticsRule ports ArabicDiacriticsRule(ResourceBundle messages).
// Java: super(messages, new Arabic());
//
//	setCategory(Categories.STYLE.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.Style);
//
// Official data is go:embed of the same diacritics.txt Java loads from the classpath.
// language is represented via LanguageCode "ar" on the base (getLocale → Locale("ar")).
func NewArabicDiacriticsRule(messages map[string]string) *ArabicDiacriticsRule {
	base := loadDiacritics()
	r := *base
	r.Messages = messages
	// Java ArabicDiacriticsRule ctor after super:
	// setCategory(Categories.STYLE.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.Style);
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	return &ArabicDiacriticsRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports ArabicDiacriticsRule.getFileNames()
// → Collections.singletonList("/ar/diacritics.txt").
func (r *ArabicDiacriticsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicDiacriticsFilePath}
}

// GetShort ports ArabicDiacriticsRule.getShort()
// → "كلمات يستحسن أن تشكّل لتصحيح نطقها".
func (r *ArabicDiacriticsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicDiacriticsRule.getMessage().
func (r *ArabicDiacriticsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports ArabicDiacriticsRule.getSuggestionsSeparator()
// → " أو " (leading space, one trailing space).
func (r *ArabicDiacriticsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Arabic always sets " أو " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports ArabicDiacriticsRule.getLocale() → Locale("ar").
func (r *ArabicDiacriticsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from diacritics.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Homophones/Redundancy/Wordiness ASR2 leaves — no rebind).
func (r *ArabicDiacriticsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
