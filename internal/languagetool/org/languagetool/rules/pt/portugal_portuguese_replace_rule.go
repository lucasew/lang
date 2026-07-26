package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via PortugalPortugueseReplaceRule(path):
//   PortugalPortuguese → /pt/pt-PT/replace.txt
//
//go:embed data/pt-PT/replace.txt
var ptPTReplaceFS embed.FS

// PortugalPortugueseSimpleReplaceID ports
// PortugalPortugueseReplaceRule.PORTUGAL_PORTUGUESE_SIMPLE_REPLACE_RULE.
const PortugalPortugueseSimpleReplaceID = "PT_PT_SIMPLE_REPLACE"

// PortugalPortugueseReplaceFilePath ports the classpath resource string
// PortugalPortuguese and tests pass to the ctor.
const PortugalPortugueseReplaceFilePath = "/pt/pt-PT/replace.txt"

// embed path for each Java classpath path (relative to this package's data/).
var ptPTReplaceEmbedByPath = map[string]string{
	PortugalPortugueseReplaceFilePath: "data/pt-PT/replace.txt",
}

var (
	ptPTReplaceMu    sync.Mutex
	ptPTReplaceBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadPTPTReplace loads the official pt-PT replace list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadPTPTReplace(path string) *rules.AbstractSimpleReplaceRule2 {
	ptPTReplaceMu.Lock()
	defer ptPTReplaceMu.Unlock()
	if base, ok := ptPTReplaceBases[path]; ok {
		return base
	}
	embedPath, ok := ptPTReplaceEmbedByPath[path]
	if !ok {
		// Java loads from classpath; unknown path is a wiring bug, not soft-skip.
		panic("PortugalPortugueseReplaceRule: unknown replace path: " + path)
	}
	f, err := ptPTReplaceFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of PortugalPortugueseReplaceRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_PT_SIMPLE_REPLACE
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.STYLE, ITSIssueType.LocaleViolation
	// - NO getUrl override (URL remains empty)
	// - isTokenException: hasPosTagStartingWith("NP") || isImmunized()
	// - example aeromoça → hospedeira de bordo
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   PortugalPortugueseSimpleReplaceID,
		Description:          "Brasileirismo: 1. palavras confundidas com as de Portugal",
		ShortMsg:             "Palavra de português do Brasil",
		MessageTemplate:      "'$match' é uma expressão brasileira, em português de Portugal utiliza-se: $suggestions",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatStyle.GetCategory(nil),
		IssueType:            rules.ITSLocaleViolation,
		IsTokenException:     ptPTReplaceTokenException,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	base.AddExamplePair(
		rules.Wrong("<marker>aeromoça</marker>"),
		rules.Fixed("<marker>hospedeira de bordo</marker>"),
	)
	ptPTReplaceBases[path] = base
	return base
}

// ptPTReplaceTokenException ports PortugalPortugueseReplaceRule.isTokenException:
// atr.hasPosTagStartingWith("NP") || atr.isImmunized().
func ptPTReplaceTokenException(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	if atr.HasPosTagStartingWith("NP") {
		return true
	}
	if atr.IsImmunized() {
		return true
	}
	return false
}

// PortugalPortugueseReplaceRule ports
// org.languagetool.rules.pt.PortugalPortugueseReplaceRule
// (extends AbstractSimpleReplaceRule2).
type PortugalPortugueseReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports PortugalPortugueseReplaceRule.path / getFileNames().
	path string
}

// NewPortugalPortugueseReplaceRule ports
// PortugalPortugueseReplaceRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewPortugalPortugueseReplaceRule(messages map[string]string, path string) *PortugalPortugueseReplaceRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("PortugalPortugueseReplaceRule: path must not be null")
	}
	base := loadPTPTReplace(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	// Java has no getUrl override — leave URL empty (do not invent).
	r.URL = ""
	r.IsTokenException = ptPTReplaceTokenException
	return &PortugalPortugueseReplaceRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports PortugalPortugueseReplaceRule.getFileNames() → singleton path.
func (r *PortugalPortugueseReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports PortugalPortugueseReplaceRule.getShort().
func (r *PortugalPortugueseReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugalPortugueseReplaceRule.getMessage().
func (r *PortugalPortugueseReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugalPortugueseReplaceRule.getSuggestionsSeparator() → " ou ".
func (r *PortugalPortugueseReplaceRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugalPortugueseReplaceRule.getLocale() → Locale("pt").
func (r *PortugalPortugueseReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *PortugalPortugueseReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
