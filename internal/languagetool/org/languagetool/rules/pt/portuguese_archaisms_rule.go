package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official archaisms lists Java loads via PortugueseArchaismsRule(path):
//   BrazilianPortuguese → /pt/pt-BR/archaisms.txt
//   PortugalPortuguese → /pt/pt-PT/archaisms.txt
//   Base Portuguese.getRelevantRules construction is commented out (#3095);
//   resource /pt/archaisms.txt still exists on the classpath.
//
//go:embed data/archaisms.txt data/pt-BR/archaisms.txt data/pt-PT/archaisms.txt
var archaismsFS embed.FS

// PortugueseArchaismsReplaceID ports PortugueseArchaismsRule.PT_ARCHAISMS_REPLACE.
const PortugueseArchaismsReplaceID = "PT_ARCHAISMS_REPLACE"

// Path constants port the classpath resource strings language modules pass to the ctor.
const (
	PortugueseArchaismsFilePath   = "/pt/archaisms.txt"
	PortugueseArchaismsBRFilePath = "/pt/pt-BR/archaisms.txt"
	PortugueseArchaismsPTFilePath = "/pt/pt-PT/archaisms.txt"
)

// PortugueseArchaismsURL ports PortugueseArchaismsRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Arca%C3%ADsmo").
const PortugueseArchaismsURL = "https://pt.wikipedia.org/wiki/Arca%C3%ADsmo"

// embed path for each Java classpath path (relative to this package's data/).
var archaismsEmbedByPath = map[string]string{
	PortugueseArchaismsFilePath:   "data/archaisms.txt",
	PortugueseArchaismsBRFilePath: "data/pt-BR/archaisms.txt",
	PortugueseArchaismsPTFilePath: "data/pt-PT/archaisms.txt",
}

var (
	archaismsMu    sync.Mutex
	archaismsBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadArchaisms loads the official archaisms list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadArchaisms(path string) *rules.AbstractSimpleReplaceRule2 {
	archaismsMu.Lock()
	defer archaismsMu.Unlock()
	if base, ok := archaismsBases[path]; ok {
		return base
	}
	embedPath, ok := archaismsEmbedByPath[path]
	if !ok {
		// Java loads from classpath; unknown path is a wiring bug, not soft-skip.
		panic("PortugueseArchaismsRule: unknown archaisms path: " + path)
	}
	f, err := archaismsFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of PortugueseArchaismsRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_ARCHAISMS_REPLACE
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.STYLE, ITSIssueType.LocaleViolation
	// - URL pt.wikipedia.org Arcaísmo
	// - NO live example pair (Java addExamplePair is commented out with câmera→câmara TODO)
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   PortugueseArchaismsReplaceID,
		Description:          "Palavras arcaicas evitáveis",
		ShortMsg:             "Arcaísmo",
		MessageTemplate:      "\"$match\" é um arcaísmo. É preferível dizer $suggestions.",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatStyle.GetCategory(nil),
		IssueType:            rules.ITSLocaleViolation,
		URL:                  PortugueseArchaismsURL,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	// Java: example pair deliberately commented out — do not invent.
	archaismsBases[path] = base
	return base
}

// PortugueseArchaismsRule ports org.languagetool.rules.pt.PortugueseArchaismsRule
// (extends AbstractSimpleReplaceRule2).
type PortugueseArchaismsRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports PortugueseArchaismsRule.path / getFileNames().
	path string
}

// NewPortugueseArchaismsRule ports
// PortugueseArchaismsRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewPortugueseArchaismsRule(messages map[string]string, path string) *PortugueseArchaismsRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("PortugueseArchaismsRule: path must not be null")
	}
	base := loadArchaisms(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	r.URL = PortugueseArchaismsURL
	return &PortugueseArchaismsRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports PortugueseArchaismsRule.getFileNames() → singleton path.
func (r *PortugueseArchaismsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports PortugueseArchaismsRule.getShort().
func (r *PortugueseArchaismsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseArchaismsRule.getMessage().
func (r *PortugueseArchaismsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseArchaismsRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseArchaismsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseArchaismsRule.getLocale() → Locale("pt").
func (r *PortugueseArchaismsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface keys from archaisms.txt).
func (r *PortugueseArchaismsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
