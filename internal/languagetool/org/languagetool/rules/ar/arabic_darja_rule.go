package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official darja list Java loads via ArabicDarjaRule:
//
//	getFileNames() → Collections.singletonList("/ar/darja.txt")
//
//go:embed data/darja.txt
var darjaFS embed.FS

// ArabicDarjaRuleID ports ArabicDarjaRule.AR_DARJA_REPLACE → "AR_DARJA_REPLACE".
const ArabicDarjaRuleID = "AR_DARJA_REPLACE"

// ArabicDarjaFilePath ports ArabicDarjaRule.FILE_NAME → "/ar/darja.txt"
// (leading slash — matches Java resource path exactly).
const ArabicDarjaFilePath = "/ar/darja.txt"

// ArabicDarjaFileNames ports ArabicDarjaRule.getFileNames().
var ArabicDarjaFileNames = []string{ArabicDarjaFilePath}

var (
	darjaOnce sync.Once
	darjaBase *rules.AbstractSimpleReplaceRule2
)

// loadDarja loads official /ar/darja.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadDarja() *rules.AbstractSimpleReplaceRule2 {
	darjaOnce.Do(func() {
		f, err := darjaFS.Open("data/darja.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicDarjaRule + AbstractSimpleReplaceRule2 defaults:
		// - id AR_DARJA_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " أو  " (space + أو + TWO trailing spaces)
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "ar" via LanguageCode (getLocale → new Locale("ar"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.STYLE (ArabicDarjaRule ctor after super)
		// - ITSIssueType.LocaleViolation (setLocQualityIssueType) — NOT Style
		// - example طرشي → فلفل حلو
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          ArabicDarjaRuleID,
			Description: "كلمات بديلة للكلمات العامية أو الأجنبية",
			// Java getShort()
			ShortMsg: "كلمات بديلة للكلمات العامية أو الأجنبية",
			// Java getMessage(): "الكلمة عامية  أو أجنبية يفضل أن يقال $suggestions"
			// (double space after عامية — preserve exact Java string)
			MessageTemplate: "الكلمة عامية  أو أجنبية يفضل أن يقال $suggestions",
			// Java getSuggestionsSeparator(): " أو  " (space + أو + two trailing spaces)
			SuggestionsSeparator: " أو  ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar", // Java: new Locale("ar")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatStyle.GetCategory(nil),
			IssueType:            rules.ITSLocaleViolation,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicDarjaFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>طرشي</marker>"),
		//                Example.fixed("<marker>فلفل حلو</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>طرشي</marker>"),
			rules.Fixed("<marker>فلفل حلو</marker>"),
		)
		darjaBase = base
	})
	return darjaBase
}

// ArabicDarjaRule ports org.languagetool.rules.ar.ArabicDarjaRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /ar/darja.txt.
type ArabicDarjaRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewArabicDarjaRule ports ArabicDarjaRule(ResourceBundle messages).
// Java: super(messages, new Arabic());
//
//	setCategory(Categories.STYLE.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.LocaleViolation);
//
// Official data is go:embed of the same darja.txt Java loads from the classpath.
// language is represented via LanguageCode "ar" on the base (getLocale → Locale("ar")).
func NewArabicDarjaRule(messages map[string]string) *ArabicDarjaRule {
	base := loadDarja()
	r := *base
	r.Messages = messages
	// Java ArabicDarjaRule ctor after super:
	// setCategory(Categories.STYLE.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.LocaleViolation);
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	return &ArabicDarjaRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports ArabicDarjaRule.getFileNames()
// → Collections.singletonList("/ar/darja.txt").
func (r *ArabicDarjaRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicDarjaFilePath}
}

// GetShort ports ArabicDarjaRule.getShort()
// → "كلمات بديلة للكلمات العامية أو الأجنبية".
func (r *ArabicDarjaRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicDarjaRule.getMessage().
func (r *ArabicDarjaRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports ArabicDarjaRule.getSuggestionsSeparator()
// → " أو  " (leading space, two trailing spaces).
func (r *ArabicDarjaRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Arabic Darja always sets " أو  " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports ArabicDarjaRule.getLocale() → Locale("ar").
func (r *ArabicDarjaRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from darja.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Diacritics/Homophones/Redundancy ASR2 leaves — no rebind).
func (r *ArabicDarjaRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
