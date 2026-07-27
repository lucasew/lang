package gl

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official wikipedia list Java loads via GalicianWikipediaRule:
//
//	getFileNames() → Collections.singletonList("/gl/wikipedia.txt")
//
//go:embed data/wikipedia.txt
var wikipediaFS embed.FS

// GalicianWikipediaCommonErrorsID ports GalicianWikipediaRule.WIKIPEDIA_COMMON_ERRORS
// → "GL_WIKIPEDIA_COMMON_ERRORS".
const GalicianWikipediaCommonErrorsID = "GL_WIKIPEDIA_COMMON_ERRORS"

// GalicianWikipediaFilePath ports GalicianWikipediaRule.FILE_NAME → "/gl/wikipedia.txt"
// (leading slash — matches Java resource path exactly).
const GalicianWikipediaFilePath = "/gl/wikipedia.txt"

// GalicianWikipediaFileNames ports GalicianWikipediaRule.getFileNames().
var GalicianWikipediaFileNames = []string{GalicianWikipediaFilePath}

// GalicianWikipediaURL ports GalicianWikipediaRule.getUrl()
// → Tools.getUrl("https://gl.wikipedia.org/wiki/Wikipedia:Erros_de_ortograf%C3%ADa_e_desviaci%C3%B3ns").
const GalicianWikipediaURL = "https://gl.wikipedia.org/wiki/Wikipedia:Erros_de_ortograf%C3%ADa_e_desviaci%C3%B3ns"

var (
	wikipediaOnce sync.Once
	wikipediaBase *rules.AbstractSimpleReplaceRule2
)

// loadWikipedia loads official /gl/wikipedia.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadWikipedia() *rules.AbstractSimpleReplaceRule2 {
	wikipediaOnce.Do(func() {
		f, err := wikipediaFS.Open("data/wikipedia.txt")
		if err != nil {
			// Java getFromRulesDirAsUrl / openStream fails closed if resource missing.
			panic(err)
		}
		defer f.Close()
		// Field twin of GalicianWikipediaRule + AbstractSimpleReplaceRule2 defaults:
		// - id GL_WIKIPEDIA_COMMON_ERRORS
		// - description / short / message as Java overrides
		// - suggestions separator " ou "
		// - CaseSensitivy.CI (base default; Java does not override isCaseSensitive/getCaseSensitivy)
		// - Locale "gl" via LanguageCode (getLocale → new Locale("gl"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.WIKIPEDIA (GalicianWikipediaRule ctor after super)
		// - ITSIssueType.Grammar (setLocQualityIssueType)
		// - URL gl.wikipedia.org Erros_de_ortografía_e_desviacións (getUrl override)
		// - example a efectos de → para os efectos de
		base := &rules.AbstractSimpleReplaceRule2{
			ID:          GalicianWikipediaCommonErrorsID,
			Description: "Erros frecuentes nos artigos da Wikipedia",
			// Java getShort()
			ShortMsg: "Erro gramatical ou de normativa",
			// Java getMessage(): "'$match' é un erro. Considere utilizar $suggestions"
			MessageTemplate: "'$match' é un erro. Considere utilizar $suggestions",
			// Java getSuggestionsSeparator(): " ou "
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "gl", // Java: new Locale("gl")
			SubRuleSpecificIDs:   false,
			Category:             rules.CatWikipedia.GetCategory(nil),
			IssueType:            rules.ITSGrammar,
			URL:                  GalicianWikipediaURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, GalicianWikipediaFilePath); err != nil {
			panic(err)
		}
		// Java:
		// addExamplePair(Example.wrong("<marker>a efectos de</marker>"),
		//                Example.fixed("<marker>para os efectos de</marker>"));
		base.AddExamplePair(
			rules.Wrong("<marker>a efectos de</marker>"),
			rules.Fixed("<marker>para os efectos de</marker>"),
		)
		wikipediaBase = base
	})
	return wikipediaBase
}

// GalicianWikipediaRule ports org.languagetool.rules.gl.GalicianWikipediaRule
// (extends AbstractSimpleReplaceRule2). Loads fixed /gl/wikipedia.txt.
type GalicianWikipediaRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewGalicianWikipediaRule ports GalicianWikipediaRule(ResourceBundle messages).
// Java: super(messages, new Galician());
//
//	setCategory(Categories.WIKIPEDIA.getCategory(messages));
//	setLocQualityIssueType(ITSIssueType.Grammar);
//
// Official data is go:embed of the same wikipedia.txt Java loads from the classpath.
// language is represented via LanguageCode "gl" on the base (getLocale → Locale("gl")).
func NewGalicianWikipediaRule(messages map[string]string) *GalicianWikipediaRule {
	base := loadWikipedia()
	r := *base
	r.Messages = messages
	// Java GalicianWikipediaRule ctor after super:
	// setCategory(Categories.WIKIPEDIA.getCategory(messages));
	// setLocQualityIssueType(ITSIssueType.Grammar);
	r.Category = rules.CatWikipedia.GetCategory(messages)
	r.IssueType = rules.ITSGrammar
	r.URL = GalicianWikipediaURL
	return &GalicianWikipediaRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports GalicianWikipediaRule.getFileNames()
// → Collections.singletonList("/gl/wikipedia.txt").
func (r *GalicianWikipediaRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{GalicianWikipediaFilePath}
}

// GetShort ports GalicianWikipediaRule.getShort() → "Erro gramatical ou de normativa".
func (r *GalicianWikipediaRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports GalicianWikipediaRule.getMessage().
func (r *GalicianWikipediaRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports GalicianWikipediaRule.getSuggestionsSeparator() → " ou ".
func (r *GalicianWikipediaRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	if r.SuggestionsSeparator == "" {
		// Java base default; Galician always sets " ou " on the twin.
		return ", "
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports GalicianWikipediaRule.getLocale() → Locale("gl").
func (r *GalicianWikipediaRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from wikipedia.txt).
// Java Match sets Rule to the concrete this; Go ASR2 base Match uses *AbstractSimpleReplaceRule2
// (same accepted pattern as Homophones/Redundancy ASR2 leaves — no rebind).
func (r *GalicianWikipediaRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
