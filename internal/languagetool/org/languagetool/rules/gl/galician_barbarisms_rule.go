package gl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official barbarisms list Java loads via GalicianBarbarismsRule:
//
//	getFileNames() → Collections.singletonList("/gl/barbarisms.txt")
//
//go:embed data/barbarisms.txt
var barbarismsFS embed.FS

// GalicianBarbarismsRuleID ports GalicianBarbarismsRule.GALICIAN_BARBARISM_RULE → "GL_BARBARISM_REPLACE".
const GalicianBarbarismsRuleID = "GL_BARBARISM_REPLACE"

// GalicianBarbarismsFilePath ports GalicianBarbarismsRule.FILE_NAME → "/gl/barbarisms.txt"
// (leading slash — matches Java resource path exactly).
const GalicianBarbarismsFilePath = "/gl/barbarisms.txt"

// GalicianBarbarismsFileNames ports GalicianBarbarismsRule.getFileNames().
var GalicianBarbarismsFileNames = []string{GalicianBarbarismsFilePath}

// GalicianBarbarismsURL ports GalicianBarbarismsRule.getUrl()
// → Tools.getUrl("https://gl.wikipedia.org/wiki/Xenismo").
const GalicianBarbarismsURL = "https://gl.wikipedia.org/wiki/Xenismo"

var (
	barbarismsOnce sync.Once
	barbarismsBase *rules.AbstractSimpleReplaceRule2
)

// loadBarbarisms loads official /gl/barbarisms.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadBarbarisms() *rules.AbstractSimpleReplaceRule2 {
	barbarismsOnce.Do(func() {
		f, err := barbarismsFS.Open("data/barbarisms.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of GalicianBarbarismsRule + AbstractSimpleReplaceRule2 defaults:
		// - id GL_BARBARISM_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " ou "
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "gl" via LanguageCode (getLocale → new Locale("gl"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.STYLE (GalicianBarbarismsRule ctor after super) — NOT REDUNDANCY
		// - ITSIssueType.LocaleViolation (setLocQualityIssueType) — NOT Style
		// - URL gl.wikipedia.org Xenismo (getUrl override)
		// - example curriculum vitae → currículo
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          GalicianBarbarismsRuleID,
			Description: "Palabras de orixe estranxeira evitábeis",
			// Java getShort()
			ShortMsg: "Xenismo",
			// Java getMessage(): "'$match' é un xenismo. É preferíbel dicir $suggestions"
			MessageTemplate: "'$match' é un xenismo. É preferíbel dicir $suggestions",
			// Java getSuggestionsSeparator(): " ou "
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl", // Java: new Locale("gl")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatStyle.GetCategory(nil),
			IssueType:            rules.ITSLocaleViolation,
			URL:                  GalicianBarbarismsURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, GalicianBarbarismsFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>curriculum vitae</marker>"),
		//                Example.fixed("<marker>currículo</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>curriculum vitae</marker>"),
			rules.Fixed("<marker>currículo</marker>"),
		)
		barbarismsBase = base
	})
	return barbarismsBase
}

// GalicianBarbarismsRule ports org.languagetool.rules.gl.GalicianBarbarismsRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /gl/barbarisms.txt.
type GalicianBarbarismsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewGalicianBarbarismsRule ports GalicianBarbarismsRule(ResourceBundle messages).
// Java: super(messages, new Galician());
//
//	setCategory(Categories.STYLE.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.LocaleViolation);
//
// Official data is go:embed of the same barbarisms.txt Java loads from the classpath.
// language is represented via LanguageCode "gl" on the base (getLocale → Locale("gl")).
func NewGalicianBarbarismsRule(messages map[string]string) *GalicianBarbarismsRule {
	base := loadBarbarisms()
	r := *base
	r.Messages = messages
	// Java GalicianBarbarismsRule ctor after super:
	// setCategory(Categories.STYLE.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.LocaleViolation);
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	r.URL = GalicianBarbarismsURL
	return &GalicianBarbarismsRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports GalicianBarbarismsRule.getFileNames()
// → Collections.singletonList("/gl/barbarisms.txt").
func (r *GalicianBarbarismsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{GalicianBarbarismsFilePath}
}

// GetShort ports GalicianBarbarismsRule.getShort() → "Xenismo".
func (r *GalicianBarbarismsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports GalicianBarbarismsRule.getMessage().
func (r *GalicianBarbarismsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports GalicianBarbarismsRule.getSuggestionsSeparator() → " ou ".
func (r *GalicianBarbarismsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Galician always sets " ou " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports GalicianBarbarismsRule.getLocale() → Locale("gl").
func (r *GalicianBarbarismsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from barbarisms.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Homophones/Redundancy ASR2 leaves — no rebind).
func (r *GalicianBarbarismsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
