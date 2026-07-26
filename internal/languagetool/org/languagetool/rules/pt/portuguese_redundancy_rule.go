package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official redundancies lists Java loads via PortugueseRedundancyRule(path):
//   Portuguese → /pt/redundancies.txt
//   BrazilianPortuguese → /pt/pt-BR/redundancies.txt
//   PortugalPortuguese → /pt/pt-PT/redundancies.txt
//
//go:embed data/redundancies.txt data/pt-BR/redundancies.txt data/pt-PT/redundancies.txt
var redundancyFS embed.FS

// PortugueseRedundancyReplaceID ports PortugueseRedundancyRule.PT_REDUNDANCY_REPLACE.
const PortugueseRedundancyReplaceID = "PT_REDUNDANCY_REPLACE"

// Path constants port the classpath resource strings language modules pass to the ctor.
const (
	PortugueseRedundancyFilePath   = "/pt/redundancies.txt"
	PortugueseRedundancyBRFilePath = "/pt/pt-BR/redundancies.txt"
	PortugueseRedundancyPTFilePath = "/pt/pt-PT/redundancies.txt"
)

// PortugueseRedundancyURL ports PortugueseRedundancyRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Pleonasmo").
const PortugueseRedundancyURL = "https://pt.wikipedia.org/wiki/Pleonasmo"

// embed path for each Java classpath path (relative to this package's data/).
var redundancyEmbedByPath = map[string]string{
	PortugueseRedundancyFilePath:   "data/redundancies.txt",
	PortugueseRedundancyBRFilePath: "data/pt-BR/redundancies.txt",
	PortugueseRedundancyPTFilePath: "data/pt-PT/redundancies.txt",
}

var (
	redundancyMu    sync.Mutex
	redundancyBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadRedundancy loads the official redundancies list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadRedundancy(path string) *rules.AbstractSimpleReplaceRule2 {
	redundancyMu.Lock()
	defer redundancyMu.Unlock()
	if base, ok := redundancyBases[path]; ok {
		return base
	}
	embedPath, ok := redundancyEmbedByPath[path]
	if !ok {
		// Java loads from classpath; unknown path is a wiring bug, not soft-skip.
		panic("PortugueseRedundancyRule: unknown redundancies path: " + path)
	}
	f, err := redundancyFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of PortugueseRedundancyRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_REDUNDANCY_REPLACE
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.REDUNDANCY, ITSIssueType.Style
	// - URL pt.wikipedia.org Pleonasmo
	// - example duna de areia → duna
	// Java getMessage() has NO trailing period after $suggestions
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   PortugueseRedundancyReplaceID,
		Description:          "1. Pleonasmos e redundâncias: $match",
		ShortMsg:             "Pleonasmo",
		MessageTemplate:      "\"$match\" é um pleonasmo. É preferível dizer $suggestions",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatRedundancy.GetCategory(nil),
		IssueType:            rules.ITSStyle,
		URL:                  PortugueseRedundancyURL,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	base.AddExamplePair(
		rules.Wrong("<marker>duna de areia</marker>"),
		rules.Fixed("<marker>duna</marker>"),
	)
	redundancyBases[path] = base
	return base
}

// PortugueseRedundancyRule ports org.languagetool.rules.pt.PortugueseRedundancyRule
// (extends AbstractSimpleReplaceRule2).
type PortugueseRedundancyRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports PortugueseRedundancyRule.path / getFileNames().
	path string
}

// NewPortugueseRedundancyRule ports
// PortugueseRedundancyRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewPortugueseRedundancyRule(messages map[string]string, path string) *PortugueseRedundancyRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("PortugueseRedundancyRule: path must not be null")
	}
	base := loadRedundancy(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatRedundancy.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.URL = PortugueseRedundancyURL
	return &PortugueseRedundancyRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports PortugueseRedundancyRule.getFileNames() → singleton path.
func (r *PortugueseRedundancyRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports PortugueseRedundancyRule.getShort().
func (r *PortugueseRedundancyRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseRedundancyRule.getMessage().
func (r *PortugueseRedundancyRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseRedundancyRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseRedundancyRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseRedundancyRule.getLocale() → Locale("pt").
func (r *PortugueseRedundancyRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from redundancies.txt).
func (r *PortugueseRedundancyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
