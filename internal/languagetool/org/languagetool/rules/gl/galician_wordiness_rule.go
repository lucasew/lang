package gl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official wordiness list Java loads via GalicianWordinessRule:
//
//	getFileNames() → Collections.singletonList("/gl/wordiness.txt")
//
//go:embed data/wordiness.txt
var wordinessFS embed.FS

// GalicianWordinessRuleID ports GalicianWordinessRule.GL_WORDINESS_REPLACE → "GL_WORDINESS_REPLACE".
const GalicianWordinessRuleID = "GL_WORDINESS_REPLACE"

// GalicianWordinessFilePath ports GalicianWordinessRule.FILE_NAME → "/gl/wordiness.txt"
// (leading slash — matches Java resource path exactly).
const GalicianWordinessFilePath = "/gl/wordiness.txt"

// GalicianWordinessFileNames ports GalicianWordinessRule.getFileNames().
var GalicianWordinessFileNames = []string{GalicianWordinessFilePath}

var (
	wordinessOnce sync.Once
	wordinessBase *rules.AbstractSimpleReplaceRule2
)

// loadWordiness loads official /gl/wordiness.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadWordiness() *rules.AbstractSimpleReplaceRule2 {
	wordinessOnce.Do(func() {
		f, err := wordinessFS.Open("data/wordiness.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of GalicianWordinessRule + AbstractSimpleReplaceRule2 defaults:
		// - id GL_WORDINESS_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " ou "
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "gl" via LanguageCode (getLocale → new Locale("gl"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.REDUNDANCY (GalicianWordinessRule ctor after super)
		// - ITSIssueType.Style (setLocQualityIssueType)
		// - getUrl is COMMENTED OUT in Java — no URL
		// - example Raramente é o caso en que acontece → Raramente acontece
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          GalicianWordinessRuleID,
			Description: "2. Expresións prolixas",
			// Java getShort()
			ShortMsg: "Expresión prolixa",
			// Java getMessage(): "'$match' é unha expresión innecesariamente complexa. É preferíbel dicir $suggestions"
			MessageTemplate: "'$match' é unha expresión innecesariamente complexa. É preferíbel dicir $suggestions",
			// Java getSuggestionsSeparator(): " ou "
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl", // Java: new Locale("gl")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatRedundancy.GetCategory(nil),
			IssueType:            rules.ITSStyle,
			// URL intentionally empty — Java getUrl is commented out
		}
		if err := base.LoadSimpleReplaceRule2Data(f, GalicianWordinessFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>Raramente é o caso en que acontece</marker> isto."),
		//                Example.fixed("<marker>Raramente acontece</marker> isto."));
		base.AddExamplePair(
			rules.Wrong("<marker>Raramente é o caso en que acontece</marker> isto."),
			rules.Fixed("<marker>Raramente acontece</marker> isto."),
		)
		wordinessBase = base
	})
	return wordinessBase
}

// GalicianWordinessRule ports org.languagetool.rules.gl.GalicianWordinessRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /gl/wordiness.txt.
type GalicianWordinessRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewGalicianWordinessRule ports GalicianWordinessRule(ResourceBundle messages).
// Java: super(messages, new Galician());
//
//	setCategory(Categories.REDUNDANCY.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.Style);
//
// Official data is go:embed of the same wordiness.txt Java loads from the classpath.
// language is represented via LanguageCode "gl" on the base (getLocale → Locale("gl")).
func NewGalicianWordinessRule(messages map[string]string) *GalicianWordinessRule {
	base := loadWordiness()
	r := *base
	r.Messages = messages
	// Java GalicianWordinessRule ctor after super:
	// setCategory(Categories.REDUNDANCY.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.Style);
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	// getUrl is commented out in Java — leave URL empty
	return &GalicianWordinessRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports GalicianWordinessRule.getFileNames()
// → Collections.singletonList("/gl/wordiness.txt").
func (r *GalicianWordinessRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{GalicianWordinessFilePath}
}

// GetShort ports GalicianWordinessRule.getShort() → "Expresión prolixa".
func (r *GalicianWordinessRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports GalicianWordinessRule.getMessage().
func (r *GalicianWordinessRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports GalicianWordinessRule.getSuggestionsSeparator() → " ou ".
func (r *GalicianWordinessRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Galician always sets " ou " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports GalicianWordinessRule.getLocale() → Locale("gl").
func (r *GalicianWordinessRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from wordiness.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Homophones/Redundancy ASR2 leaves — no rebind).
func (r *GalicianWordinessRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
