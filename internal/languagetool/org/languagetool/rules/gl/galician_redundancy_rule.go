package gl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official redundancy list Java loads via GalicianRedundancyRule:
//
//	getFileNames() → Collections.singletonList("/gl/redundancies.txt")
//
//go:embed data/redundancies.txt
var redundancyFS embed.FS

// GalicianRedundancyRuleID ports GalicianRedundancyRule.GL_REDUNDANCY_REPLACE → "GL_REDUNDANCY_REPLACE".
const GalicianRedundancyRuleID = "GL_REDUNDANCY_REPLACE"

// GalicianRedundancyFilePath ports GalicianRedundancyRule.FILE_NAME → "/gl/redundancies.txt"
// (leading slash — matches Java resource path exactly).
const GalicianRedundancyFilePath = "/gl/redundancies.txt"

// GalicianRedundancyFileNames ports GalicianRedundancyRule.getFileNames().
var GalicianRedundancyFileNames = []string{GalicianRedundancyFilePath}

// GalicianRedundancyURL ports GalicianRedundancyRule.getUrl()
// → Tools.getUrl("https://gl.wikipedia.org/wiki/Pleonasmo").
const GalicianRedundancyURL = "https://gl.wikipedia.org/wiki/Pleonasmo"

var (
	redundancyOnce sync.Once
	redundancyBase *rules.AbstractSimpleReplaceRule2
)

// loadRedundancy loads official /gl/redundancies.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadRedundancy() *rules.AbstractSimpleReplaceRule2 {
	redundancyOnce.Do(func() {
		f, err := redundancyFS.Open("data/redundancies.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of GalicianRedundancyRule + AbstractSimpleReplaceRule2 defaults:
		// - id GL_REDUNDANCY_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " ou "
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "gl" via LanguageCode (getLocale → new Locale("gl"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.REDUNDANCY (GalicianRedundancyRule ctor after super)
		// - ITSIssueType.Style (setLocQualityIssueType)
		// - URL gl.wikipedia.org Pleonasmo (getUrl override)
		// - example duna de area → duna
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          GalicianRedundancyRuleID,
			Description: "1. Pleonasmos e redundancias",
			// Java getShort()
			ShortMsg: "Pleonasmo",
			// Java getMessage(): "'$match' é un pleonasmo. É preferible dicir $suggestions"
			MessageTemplate: "'$match' é un pleonasmo. É preferible dicir $suggestions",
			// Java getSuggestionsSeparator(): " ou "
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl", // Java: new Locale("gl")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatRedundancy.GetCategory(nil),
			IssueType:            rules.ITSStyle,
			URL:                  GalicianRedundancyURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, GalicianRedundancyFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>duna de area</marker>"),
		//                Example.fixed("<marker>duna</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>duna de area</marker>"),
			rules.Fixed("<marker>duna</marker>"),
		)
		redundancyBase = base
	})
	return redundancyBase
}

// GalicianRedundancyRule ports org.languagetool.rules.gl.GalicianRedundancyRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /gl/redundancies.txt.
type GalicianRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewGalicianRedundancyRule ports GalicianRedundancyRule(ResourceBundle messages).
// Java: super(messages, new Galician());
//
//	setCategory(Categories.REDUNDANCY.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.Style);
//
// Official data is go:embed of the same redundancies.txt Java loads from the classpath.
// language is represented via LanguageCode "gl" on the base (getLocale → Locale("gl")).
func NewGalicianRedundancyRule(messages map[string]string) *GalicianRedundancyRule {
	base := loadRedundancy()
	r := *base
	r.Messages = messages
	// Java GalicianRedundancyRule ctor after super:
	// setCategory(Categories.REDUNDANCY.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.Style);
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.URL = GalicianRedundancyURL
	return &GalicianRedundancyRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports GalicianRedundancyRule.getFileNames()
// → Collections.singletonList("/gl/redundancies.txt").
func (r *GalicianRedundancyRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{GalicianRedundancyFilePath}
}

// GetShort ports GalicianRedundancyRule.getShort() → "Pleonasmo".
func (r *GalicianRedundancyRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports GalicianRedundancyRule.getMessage().
func (r *GalicianRedundancyRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports GalicianRedundancyRule.getSuggestionsSeparator() → " ou ".
func (r *GalicianRedundancyRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Galician always sets " ou " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports GalicianRedundancyRule.getLocale() → Locale("gl").
func (r *GalicianRedundancyRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from redundancies.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Homophones/Greek ASR2 leaves — no rebind).
func (r *GalicianRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
