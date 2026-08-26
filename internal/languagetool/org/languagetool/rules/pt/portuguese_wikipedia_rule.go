package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official wikipedia lists Java loads via PortugueseWikipediaRule(path):
//   Portuguese → /pt/wikipedia.txt
//   BrazilianPortuguese → /pt/pt-BR/wikipedia.txt
//   PortugalPortuguese → /pt/pt-PT/wikipedia.txt
//
//go:embed data/wikipedia.txt data/pt-BR/wikipedia.txt data/pt-PT/wikipedia.txt
var wikipediaFS embed.FS

// PortugueseWikipediaCommonErrorsID ports PortugueseWikipediaRule.WIKIPEDIA_COMMON_ERRORS.
const PortugueseWikipediaCommonErrorsID = "PT_WIKIPEDIA_COMMON_ERRORS"

// Path constants port the classpath resource strings language modules pass to the ctor.
const (
	PortugueseWikipediaFilePath   = "/pt/wikipedia.txt"
	PortugueseWikipediaBRFilePath = "/pt/pt-BR/wikipedia.txt"
	PortugueseWikipediaPTFilePath = "/pt/pt-PT/wikipedia.txt"
)

// PortugueseWikipediaURL ports PortugueseWikipediaRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Wikip%C3%A9dia:Lista_de_erros_comuns/M%C3%A1quinas").
const PortugueseWikipediaURL = "https://pt.wikipedia.org/wiki/Wikip%C3%A9dia:Lista_de_erros_comuns/M%C3%A1quinas"

// embed path for each Java classpath path (relative to this package's data/).
var wikipediaEmbedByPath = map[string]string{
	PortugueseWikipediaFilePath:   "data/wikipedia.txt",
	PortugueseWikipediaBRFilePath: "data/pt-BR/wikipedia.txt",
	PortugueseWikipediaPTFilePath: "data/pt-PT/wikipedia.txt",
}

var (
	wikipediaMu    sync.Mutex
	wikipediaBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadWikipedia loads the official wikipedia list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadWikipedia(path string) *rules.AbstractSimpleReplaceRule2 {
	wikipediaMu.Lock()
	defer wikipediaMu.Unlock()
	if base, ok := wikipediaBases[path]; ok {
		return base
	}
	embedPath, ok := wikipediaEmbedByPath[path]
	if !ok {
		// Java loads from classpath; unknown path is a wiring bug, not soft-skip.
		panic("PortugueseWikipediaRule: unknown wikipedia path: " + path)
	}
	f, err := wikipediaFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of PortugueseWikipediaRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_WIKIPEDIA_COMMON_ERRORS
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.WIKIPEDIA, ITSIssueType.Grammar
	// - URL pt.wikipedia.org Lista_de_erros_comuns/Máquinas
	// - NO example pair (Java addExamplePair is commented out)
	// Java getMessage() has NO trailing period after $suggestions
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   PortugueseWikipediaCommonErrorsID,
		Description:          "Erros frequentes nos artigos da Wikipédia: $match",
		ShortMsg:             "Erro gramatical ou de normativa",
		MessageTemplate:      "Possível erro em \"$match\". Prefira $suggestions",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatWikipedia.GetCategory(nil),
		IssueType:            rules.ITSGrammar,
		URL:                  PortugueseWikipediaURL,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	// Java: example pair deliberately removed / commented out — do not invent.
	wikipediaBases[path] = base
	return base
}

// PortugueseWikipediaRule ports org.languagetool.rules.pt.PortugueseWikipediaRule
// (extends AbstractSimpleReplaceRule2).
type PortugueseWikipediaRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports PortugueseWikipediaRule.path / getFileNames().
	path string
}

// NewPortugueseWikipediaRule ports
// PortugueseWikipediaRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewPortugueseWikipediaRule(messages map[string]string, path string) *PortugueseWikipediaRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("PortugueseWikipediaRule: path must not be null")
	}
	base := loadWikipedia(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatWikipedia.GetCategory(messages)
	r.IssueType = rules.ITSGrammar
	r.URL = PortugueseWikipediaURL
	return &PortugueseWikipediaRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports PortugueseWikipediaRule.getFileNames() → singleton path.
func (r *PortugueseWikipediaRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports PortugueseWikipediaRule.getShort().
func (r *PortugueseWikipediaRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseWikipediaRule.getMessage().
func (r *PortugueseWikipediaRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseWikipediaRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseWikipediaRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseWikipediaRule.getLocale() → Locale("pt").
func (r *PortugueseWikipediaRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from wikipedia.txt).
func (r *PortugueseWikipediaRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
