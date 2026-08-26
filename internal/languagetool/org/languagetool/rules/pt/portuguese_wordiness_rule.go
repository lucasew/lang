package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official wordiness lists Java loads via PortugueseWordinessRule(path):
//   Portuguese → /pt/wordiness.txt
//   BrazilianPortuguese → /pt/pt-BR/wordiness.txt
//   PortugalPortuguese → /pt/pt-PT/wordiness.txt
//
//go:embed data/wordiness.txt data/pt-BR/wordiness.txt data/pt-PT/wordiness.txt
var wordinessFS embed.FS

// PortugueseWordinessReplaceID ports PortugueseWordinessRule.PT_WORDINESS_REPLACE.
const PortugueseWordinessReplaceID = "PT_WORDINESS_REPLACE"

// Path constants port the classpath resource strings language modules pass to the ctor.
const (
	PortugueseWordinessFilePath   = "/pt/wordiness.txt"
	PortugueseWordinessBRFilePath = "/pt/pt-BR/wordiness.txt"
	PortugueseWordinessPTFilePath = "/pt/pt-PT/wordiness.txt"
)

// PortugueseWordinessURL ports PortugueseWordinessRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/V%C3%ADcio_de_linguagem#Prolixidade_ou_preciosismo").
const PortugueseWordinessURL = "https://pt.wikipedia.org/wiki/V%C3%ADcio_de_linguagem#Prolixidade_ou_preciosismo"

// embed path for each Java classpath path (relative to this package's data/).
var wordinessEmbedByPath = map[string]string{
	PortugueseWordinessFilePath:   "data/wordiness.txt",
	PortugueseWordinessBRFilePath: "data/pt-BR/wordiness.txt",
	PortugueseWordinessPTFilePath: "data/pt-PT/wordiness.txt",
}

var (
	wordinessMu    sync.Mutex
	wordinessBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadWordiness loads the official wordiness list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadWordiness(path string) *rules.AbstractSimpleReplaceRule2 {
	wordinessMu.Lock()
	defer wordinessMu.Unlock()
	if base, ok := wordinessBases[path]; ok {
		return base
	}
	embedPath, ok := wordinessEmbedByPath[path]
	if !ok {
		// Java loads from classpath; unknown path is a wiring bug, not soft-skip.
		panic("PortugueseWordinessRule: unknown wordiness path: " + path)
	}
	f, err := wordinessFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of PortugueseWordinessRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_WORDINESS_REPLACE
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.REDUNDANCY, ITSIssueType.Style
	// - URL pt.wikipedia.org Prolixidade_ou_preciosismo
	// - example Raramente é o caso em que acontece → Raramente acontece
	// Java getMessage() HAS trailing period after $suggestions (unlike Redundancy)
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   PortugueseWordinessReplaceID,
		Description:          "2. Expressões prolixas: $match",
		ShortMsg:             "Expressão prolixa",
		MessageTemplate:      "\"$match\" é uma expressão prolixa. É preferível dizer $suggestions.",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatRedundancy.GetCategory(nil),
		IssueType:            rules.ITSStyle,
		URL:                  PortugueseWordinessURL,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	base.AddExamplePair(
		rules.Wrong("<marker>Raramente é o caso em que acontece</marker> isto."),
		rules.Fixed("<marker>Raramente acontece</marker> isto."),
	)
	wordinessBases[path] = base
	return base
}

// PortugueseWordinessRule ports org.languagetool.rules.pt.PortugueseWordinessRule
// (extends AbstractSimpleReplaceRule2).
type PortugueseWordinessRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports PortugueseWordinessRule.path / getFileNames().
	path string
}

// NewPortugueseWordinessRule ports
// PortugueseWordinessRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewPortugueseWordinessRule(messages map[string]string, path string) *PortugueseWordinessRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("PortugueseWordinessRule: path must not be null")
	}
	base := loadWordiness(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.URL = PortugueseWordinessURL
	return &PortugueseWordinessRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports PortugueseWordinessRule.getFileNames() → singleton path.
func (r *PortugueseWordinessRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports PortugueseWordinessRule.getShort().
func (r *PortugueseWordinessRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseWordinessRule.getMessage().
func (r *PortugueseWordinessRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseWordinessRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseWordinessRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseWordinessRule.getLocale() → Locale("pt").
func (r *PortugueseWordinessRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from wordiness.txt).
func (r *PortugueseWordinessRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
