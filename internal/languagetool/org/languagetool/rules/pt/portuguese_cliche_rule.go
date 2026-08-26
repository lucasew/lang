package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official cliches lists Java loads via PortugueseClicheRule(path):
//
//	Portuguese → /pt/cliches.txt
//	BrazilianPortuguese → /pt/pt-BR/cliches.txt
//	PortugalPortuguese → /pt/pt-PT/cliches.txt
//
//go:embed data/cliches.txt data/pt-BR/cliches.txt data/pt-PT/cliches.txt
var clichesFS embed.FS

// PortugueseClicheReplaceID ports PortugueseClicheRule.PORTUGUESE_CLICHE_RULE.
const PortugueseClicheReplaceID = "PT_CLICHE_REPLACE"

// Path constants port the classpath resource strings language modules pass to the ctor.
const (
	PortugueseClicheFilePath   = "/pt/cliches.txt"
	PortugueseClicheBRFilePath = "/pt/pt-BR/cliches.txt"
	PortugueseClichePTFilePath = "/pt/pt-PT/cliches.txt"
)

// PortugueseClicheURL ports PortugueseClicheRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Clichê").
const PortugueseClicheURL = "https://pt.wikipedia.org/wiki/Clichê"

// embed path for each Java classpath path (relative to this package's data/).
var clicheEmbedByPath = map[string]string{
	PortugueseClicheFilePath:   "data/cliches.txt",
	PortugueseClicheBRFilePath: "data/pt-BR/cliches.txt",
	PortugueseClichePTFilePath: "data/pt-PT/cliches.txt",
}

var (
	clicheMu    sync.Mutex
	clicheBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadCliche loads the official cliches list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadCliche(path string) *rules.AbstractSimpleReplaceRule2 {
	clicheMu.Lock()
	defer clicheMu.Unlock()
	if base, ok := clicheBases[path]; ok {
		return base
	}
	embedPath, ok := clicheEmbedByPath[path]
	if !ok {
		// Java would fail to open unknown classpath path — fail closed.
		panic("PortugueseClicheRule: unknown cliches path: " + path)
	}
	f, err := clichesFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of PortugueseClicheRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_CLICHE_REPLACE
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.STYLE, ITSIssueType.Style
	// - URL pt.wikipedia.org Clichê
	// - example quente como uma fornalha → quente
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   PortugueseClicheReplaceID,
		Description:          "Frases-feitas e expressões idiomáticas: $match",
		ShortMsg:             "Frase-feita",
		MessageTemplate:      "\"$match\" é uma frase-feita. É preferível dizer $suggestions.",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatStyle.GetCategory(nil),
		IssueType:            rules.ITSStyle,
		URL:                  PortugueseClicheURL,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	base.AddExamplePair(
		rules.Wrong("<marker>quente como uma fornalha</marker>"),
		rules.Fixed("<marker>quente</marker>"),
	)
	clicheBases[path] = base
	return base
}

// PortugueseClicheRule ports org.languagetool.rules.pt.PortugueseClicheRule
// (extends AbstractSimpleReplaceRule2).
type PortugueseClicheRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports PortugueseClicheRule.path / getFileNames().
	path string
}

// NewPortugueseClicheRule ports PortugueseClicheRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewPortugueseClicheRule(messages map[string]string, path string) *PortugueseClicheRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("PortugueseClicheRule: path must not be null")
	}
	base := loadCliche(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.URL = PortugueseClicheURL
	return &PortugueseClicheRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports PortugueseClicheRule.getFileNames() → singleton path.
func (r *PortugueseClicheRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports PortugueseClicheRule.getShort().
func (r *PortugueseClicheRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseClicheRule.getMessage().
func (r *PortugueseClicheRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseClicheRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseClicheRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseClicheRule.getLocale() → Locale("pt").
func (r *PortugueseClicheRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from cliches.txt).
func (r *PortugueseClicheRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
