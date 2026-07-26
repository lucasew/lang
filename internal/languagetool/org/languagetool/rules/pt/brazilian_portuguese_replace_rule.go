package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Official replace list Java loads via BrazilianPortugueseReplaceRule(path):
//   BrazilianPortuguese → /pt/pt-BR/replace.txt
//
//go:embed data/pt-BR/replace.txt
var brReplaceFS embed.FS

// BrazilianPortugueseSimpleReplaceID ports
// BrazilianPortugueseReplaceRule.BRAZILIAN_PORTUGUESE_SIMPLE_REPLACE_RULE.
const BrazilianPortugueseSimpleReplaceID = "PT_BR_SIMPLE_REPLACE"

// BrazilianPortugueseReplaceFilePath ports the classpath resource string
// BrazilianPortuguese and BrazilianPortugueseSimpleReplaceRuleTest pass to the ctor.
const BrazilianPortugueseReplaceFilePath = "/pt/pt-BR/replace.txt"

// embed path for each Java classpath path (relative to this package's data/).
var brReplaceEmbedByPath = map[string]string{
	BrazilianPortugueseReplaceFilePath: "data/pt-BR/replace.txt",
}

var (
	brReplaceMu    sync.Mutex
	brReplaceBases = map[string]*rules.AbstractSimpleReplaceRule2{}
)

// loadBRReplace loads the official pt-BR replace list for path once
// (Java AbstractSimpleReplaceRule2.lazyInit → fillMaps over getFileNames()).
func loadBRReplace(path string) *rules.AbstractSimpleReplaceRule2 {
	brReplaceMu.Lock()
	defer brReplaceMu.Unlock()
	if base, ok := brReplaceBases[path]; ok {
		return base
	}
	embedPath, ok := brReplaceEmbedByPath[path]
	if !ok {
		// Java loads from classpath; unknown path is a wiring bug, not soft-skip.
		panic("BrazilianPortugueseReplaceRule: unknown replace path: " + path)
	}
	f, err := brReplaceFS.Open(embedPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Field twin of BrazilianPortugueseReplaceRule + AbstractSimpleReplaceRule2 defaults:
	// - id PT_BR_SIMPLE_REPLACE
	// - description / short / message as Java overrides
	// - suggestions separator " ou "
	// - CaseSensitivy.CI (base default)
	// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
	// - useSubRuleSpecificIds()
	// - Categories.STYLE, ITSIssueType.LocaleViolation
	// - NO getUrl override (URL remains empty)
	// - isTokenException: hasPosTagStartingWith("NP") || isImmunized()
	// - example hospedeira de bordo → aeromoça
	base := &rules.AbstractSimpleReplaceRule2{
		ID:                   BrazilianPortugueseSimpleReplaceID,
		Description:          "Palavras portuguesas facilmente confundidas com as do Brasil",
		ShortMsg:             "Palavra do português de Portugal",
		MessageTemplate:      "\"$match\" é uma expressão usada sobretudo em Portugal. No português brasileiro diz-se $suggestions",
		SuggestionsSeparator: " ou ",
		CaseSens:             rules.CaseInsensitive,
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Category:             rules.CatStyle.GetCategory(nil),
		IssueType:            rules.ITSLocaleViolation,
		IsTokenException:     brReplaceTokenException,
	}
	if err := base.LoadSimpleReplaceRule2Data(f, path); err != nil {
		panic(err)
	}
	base.AddExamplePair(
		rules.Wrong("<marker>hospedeira de bordo</marker>"),
		rules.Fixed("<marker>aeromoça</marker>"),
	)
	brReplaceBases[path] = base
	return base
}

// brReplaceTokenException ports BrazilianPortugueseReplaceRule.isTokenException:
// atr.hasPosTagStartingWith("NP") || atr.isImmunized().
func brReplaceTokenException(atr *languagetool.AnalyzedTokenReadings) bool {
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

// BrazilianPortugueseReplaceRule ports
// org.languagetool.rules.pt.BrazilianPortugueseReplaceRule
// (extends AbstractSimpleReplaceRule2).
type BrazilianPortugueseReplaceRule struct {
	*rules.AbstractSimpleReplaceRule2
	// path ports BrazilianPortugueseReplaceRule.path / getFileNames().
	path string
}

// NewBrazilianPortugueseReplaceRule ports
// BrazilianPortugueseReplaceRule(ResourceBundle messages, String path, Language language).
// path is Objects.requireNonNull; language is represented via LanguageCode "pt"
// on the base (getLocale → Locale("pt")).
func NewBrazilianPortugueseReplaceRule(messages map[string]string, path string) *BrazilianPortugueseReplaceRule {
	if path == "" {
		// Java: Objects.requireNonNull(path)
		panic("BrazilianPortugueseReplaceRule: path must not be null")
	}
	base := loadBRReplace(path)
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSLocaleViolation
	// Java has no getUrl override — leave URL empty (do not invent).
	r.URL = ""
	r.IsTokenException = brReplaceTokenException
	return &BrazilianPortugueseReplaceRule{AbstractSimpleReplaceRule2: &r, path: path}
}

// GetFileNames ports BrazilianPortugueseReplaceRule.getFileNames() → singleton path.
func (r *BrazilianPortugueseReplaceRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{r.path}
}

// GetShort ports BrazilianPortugueseReplaceRule.getShort().
func (r *BrazilianPortugueseReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports BrazilianPortugueseReplaceRule.getMessage().
func (r *BrazilianPortugueseReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports BrazilianPortugueseReplaceRule.getSuggestionsSeparator() → " ou ".
func (r *BrazilianPortugueseReplaceRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports BrazilianPortugueseReplaceRule.getLocale() → Locale("pt").
func (r *BrazilianPortugueseReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from replace.txt).
func (r *BrazilianPortugueseReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
