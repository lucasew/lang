package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official redundancy list Java loads via ArabicRedundancyRule:
//
//	getFileNames() → Collections.singletonList("/ar/redundancies.txt")
//
//go:embed data/redundancies.txt
var redundancyFS embed.FS

// ArabicRedundancyRuleID ports ArabicRedundancyRule.AR_REDUNDANCY_REPLACE → "AR_REDUNDANCY_REPLACE".
const ArabicRedundancyRuleID = "AR_REDUNDANCY_REPLACE"

// ArabicRedundancyFilePath ports ArabicRedundancyRule.FILE_NAME → "/ar/redundancies.txt"
// (leading slash — matches Java resource path exactly).
const ArabicRedundancyFilePath = "/ar/redundancies.txt"

// ArabicRedundancyFileNames ports ArabicRedundancyRule.getFileNames().
var ArabicRedundancyFileNames = []string{ArabicRedundancyFilePath}

var (
	redundancyOnce sync.Once
	redundancyBase *rules.AbstractSimpleReplaceRule2
)

// loadRedundancy loads official /ar/redundancies.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadRedundancy() *rules.AbstractSimpleReplaceRule2 {
	redundancyOnce.Do(func() {
		f, err := redundancyFS.Open("data/redundancies.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of ArabicRedundancyRule + AbstractSimpleReplaceRule2 defaults:
		// - id AR_REDUNDANCY_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " أو " (space + أو + one trailing space)
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "ar" via LanguageCode (getLocale → new Locale("ar"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.REDUNDANCY (ArabicRedundancyRule ctor after super)
		// - ITSIssueType.Style (setLocQualityIssueType)
		// - example سوف لن → لن
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          ArabicRedundancyRuleID,
			Description: "1. تكرار (عام)",
			// Java getShort()
			ShortMsg: "تكرار",
			// Java getMessage(): "'$match' تعبير فيه تكرار.في بعض الحالات، يستحسن استعمال $suggestions"
			MessageTemplate: "'$match' تعبير فيه تكرار.في بعض الحالات، يستحسن استعمال $suggestions",
			// Java getSuggestionsSeparator(): " أو " (space + أو + one trailing space)
			SuggestionsSeparator: " أو ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ar", // Java: new Locale("ar")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatRedundancy.GetCategory(nil),
			IssueType:            rules.ITSStyle,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, ArabicRedundancyFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>سوف لن</marker>"),
		//                Example.fixed("<marker>لن</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>سوف لن</marker>"),
			rules.Fixed("<marker>لن</marker>"),
		)
		redundancyBase = base
	})
	return redundancyBase
}

// ArabicRedundancyRule ports org.languagetool.rules.ar.ArabicRedundancyRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /ar/redundancies.txt.
type ArabicRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewArabicRedundancyRule ports ArabicRedundancyRule(ResourceBundle messages).
// Java: super(messages, new Arabic());
//
//	super.setCategory(Categories.REDUNDANCY.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.Style);
//
// Official data is go:embed of the same redundancies.txt Java loads from the classpath.
// language is represented via LanguageCode "ar" on the base (getLocale → Locale("ar")).
func NewArabicRedundancyRule(messages map[string]string) *ArabicRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	// Java ArabicRedundancyRule ctor after super:
	// super.setCategory(Categories.REDUNDANCY.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.Style);
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	return &ArabicRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports ArabicRedundancyRule.getFileNames()
// → Collections.singletonList("/ar/redundancies.txt").
func (r *ArabicRedundancyRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{ArabicRedundancyFilePath}
}

// GetShort ports ArabicRedundancyRule.getShort() → "تكرار".
func (r *ArabicRedundancyRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports ArabicRedundancyRule.getMessage().
func (r *ArabicRedundancyRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports ArabicRedundancyRule.getSuggestionsSeparator()
// → " أو " (leading space, one trailing space).
func (r *ArabicRedundancyRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Arabic always sets " أو " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports ArabicRedundancyRule.getLocale() → Locale("ar").
func (r *ArabicRedundancyRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from redundancies.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Homophones/Greek ASR2 leaves — no rebind).
func (r *ArabicRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
