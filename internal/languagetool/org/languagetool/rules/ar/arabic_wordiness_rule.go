package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official wordiness list Java loads via ArabicWordinessRule:
//
//	getFileNames() → Collections.singletonList("/ar/wordiness.txt")
//
//go:embed data/wordiness.txt
var wordinessFS embed.FS

// ArabicWordinessRuleID ports ArabicWordinessRule.AR_WORDINESS_REPLACE → "AR_WORDINESS_REPLACE".
const ArabicWordinessRuleID = "AR_WORDINESS_REPLACE"

// ArabicWordinessFilePath ports ArabicWordinessRule.FILE_NAME → "/ar/wordiness.txt"
// (leading slash — matches Java resource path exactly).
const ArabicWordinessFilePath = "/ar/wordiness.txt"

// ArabicWordinessFileNames ports ArabicWordinessRule.getFileNames().
var ArabicWordinessFileNames = []string{ArabicWordinessFilePath}

var (
	wordinessOnce sync.Once
	wordinessBase *rules.AbstractSimpleReplaceRule2
)

// loadWordiness loads official /ar/wordiness.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadWordiness() *rules.AbstractSimpleReplaceRule2 {
	wordinessOnce.Do(func() {
		f, err := wordinessFS.Open("data/wordiness.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicWordinessRule + AbstractSimpleReplaceRule2 defaults:
		// - id AR_WORDINESS_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " أو " (space + أو + one trailing space)
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "ar" via LanguageCode (getLocale → new Locale("ar"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.REDUNDANCY (ArabicWordinessRule ctor after super)
		// - ITSIssueType.Style (setLocQualityIssueType)
		// - example هناك خطأ في العبارة → في العبارة خطأ
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          ArabicWordinessRuleID,
			Description: "2. حشو(تعبير فيه تكرار)",
			// Java getShort()
			ShortMsg: "حشو (تعبير فيه تكرار)",
			// Java getMessage(): "'$match' تعبير فيه حشو يفضل أن يقال $suggestions"
			MessageTemplate: "'$match' تعبير فيه حشو يفضل أن يقال $suggestions",
			// Java getSuggestionsSeparator(): " أو " (space + أو + one trailing space)
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar", // Java: new Locale("ar")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatRedundancy.GetCategory(nil),
			IssueType:            rules.ITSStyle,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicWordinessFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>هناك خطأ في العبارة</marker>"),
		//                Example.fixed("<marker>في العبارة خطأ</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>هناك خطأ في العبارة</marker>"),
			rules.Fixed("<marker>في العبارة خطأ</marker>"),
		)
		wordinessBase = base
	})
	return wordinessBase
}

// ArabicWordinessRule ports org.languagetool.rules.ar.ArabicWordinessRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /ar/wordiness.txt.
type ArabicWordinessRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewArabicWordinessRule ports ArabicWordinessRule(ResourceBundle messages).
// Java: super(messages, new Arabic());
//
//	super.setCategory(Categories.REDUNDANCY.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.Style);
//
// Official data is go:embed of the same wordiness.txt Java loads from the classpath.
// language is represented via LanguageCode "ar" on the base (getLocale → Locale("ar")).
func NewArabicWordinessRule(messages map[string]string) *ArabicWordinessRule {
	base := loadWordiness()
	r := *base
	r.Messages = messages
	// Java ArabicWordinessRule ctor after super:
	// setCategory(Categories.REDUNDANCY.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.Style);
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	return &ArabicWordinessRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports ArabicWordinessRule.getFileNames()
// → Collections.singletonList("/ar/wordiness.txt").
func (r *ArabicWordinessRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicWordinessFilePath}
}

// GetShort ports ArabicWordinessRule.getShort() → "حشو (تعبير فيه تكرار)".
func (r *ArabicWordinessRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicWordinessRule.getMessage().
func (r *ArabicWordinessRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports ArabicWordinessRule.getSuggestionsSeparator()
// → " أو " (leading space, one trailing space).
func (r *ArabicWordinessRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Arabic always sets " أو " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports ArabicWordinessRule.getLocale() → Locale("ar").
func (r *ArabicWordinessRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from wordiness.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Homophones/Redundancy ASR2 leaves — no rebind).
func (r *ArabicWordinessRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
