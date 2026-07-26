package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official barbarisms lists Java loads via PortugueseBarbarismsRule(path):
//   Portuguese → /pt/barbarisms.txt
//   BrazilianPortuguese → /pt/pt-BR/barbarisms.txt
//   PortugalPortuguese → /pt/pt-PT/barbarisms.txt
//
//go:embed data/barbarisms.txt data/pt-BR/barbarisms.txt data/pt-PT/barbarisms.txt
var barbarismsFS embed.FS

// PortugueseBarbarismsReplaceID ports PortugueseBarbarismsRule.PT_BARBARISMS_REPLACE.
const PortugueseBarbarismsReplaceID = "PT_BARBARISMS_REPLACE"

// Path constants port the classpath resource strings language modules pass to the ctor.
const (
	PortugueseBarbarismsFilePath   = "/pt/barbarisms.txt"
	PortugueseBarbarismsBRFilePath = "/pt/pt-BR/barbarisms.txt"
	PortugueseBarbarismsPTFilePath = "/pt/pt-PT/barbarisms.txt"
)

// PortugueseBarbarismsURL ports PortugueseBarbarismsRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Estrangeirismo").
const PortugueseBarbarismsURL = "https://pt.wikipedia.org/wiki/Estrangeirismo"

// embed path for each Java classpath path (relative to this package's data/).
var barbarismsEmbedByPath = map[string]string{
	PortugueseBarbarismsFilePath:   "data/barbarisms.txt",
	PortugueseBarbarismsBRFilePath: "data/pt-BR/barbarisms.txt",
	PortugueseBarbarismsPTFilePath: "data/pt-PT/barbarisms.txt",
}

var (
	barbarismsMu    sync.Mutex
	barbarismsBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadBarbarisms loads the official barbarisms list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadBarbarisms(path string) *rules.AbstractSimpleReplaceRule2 {
	barbarismsMu.Lock()
	defer barbarismsMu.Unlock()
	if base, ok := barbarismsBases[path]; ok {
		return base
	}
	embedPath, ok := barbarismsEmbedByPath[path]
	if !ok {
		// Java loads from classpath; unknown path is a wiring bug, not soft-skip.
		panic("PortugueseBarbarismsRule: unknown barbarisms path: " + path)
	}
	f, err := barbarismsFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of PortugueseBarbarismsRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_BARBARISMS_REPLACE
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.STYLE, ITSIssueType.LocaleViolation (not Style)
	// - URL pt.wikipedia.org Estrangeirismo
	// - isTokenException: immunized || NP* || _english_ignore_*
	// - example curriculum vitae → currículo
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   PortugueseBarbarismsReplaceID,
		Description:          "Palavras de origem estrangeira evitáveis: $match",
		ShortMsg:             "Estrangeirismo",
		MessageTemplate:      "\"$match\" é um estrangeirismo. É preferível dizer $suggestions.",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatStyle.GetCategory(nil),
		IssueType:            rules.ITSLocaleViolation,
		URL:                  PortugueseBarbarismsURL,
		IsTokenException:     barbarismsTokenException,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	base.AddExamplePair(
		rules.Wrong("<marker>curriculum vitae</marker>"),
		rules.Fixed("<marker>currículo</marker>"),
	)
	barbarismsBases[path] = base
	return base
}

// barbarismsTokenException ports PortugueseBarbarismsRule.isTokenException:
// atr.isImmunized() || atr.hasPosTagStartingWith("NP") || atr.hasPosTagStartingWith("_english_ignore_").
func barbarismsTokenException(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	if atr.IsImmunized() {
		return true
	}
	if atr.HasPosTagStartingWith("NP") {
		return true
	}
	if atr.HasPosTagStartingWith("_english_ignore_") {
		return true
	}
	return false
}

// PortugueseBarbarismsRule ports org.languagetool.rules.pt.PortugueseBarbarismsRule
// (extends AbstractSimpleReplaceRule2).
type PortugueseBarbarismsRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports PortugueseBarbarismsRule.path / getFileNames().
	path string
}

// NewPortugueseBarbarismsRule ports
// PortugueseBarbarismsRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewPortugueseBarbarismsRule(messages map[string]string, path string) *PortugueseBarbarismsRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("PortugueseBarbarismsRule: path must not be null")
	}
	base := loadBarbarisms(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	r.URL = PortugueseBarbarismsURL
	r.IsTokenException = barbarismsTokenException
	return &PortugueseBarbarismsRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports PortugueseBarbarismsRule.getFileNames() → singleton path.
func (r *PortugueseBarbarismsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports PortugueseBarbarismsRule.getShort().
func (r *PortugueseBarbarismsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseBarbarismsRule.getMessage().
func (r *PortugueseBarbarismsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseBarbarismsRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseBarbarismsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseBarbarismsRule.getLocale() → Locale("pt").
func (r *PortugueseBarbarismsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from barbarisms.txt).
func (r *PortugueseBarbarismsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
