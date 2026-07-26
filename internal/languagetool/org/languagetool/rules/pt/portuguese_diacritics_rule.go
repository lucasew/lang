package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/diacritics.txt
var diacriticsFS embed.FS

// PortugueseDiacriticsReplaceID ports PortugueseDiacriticsRule.PT_DIACRITICS_REPLACE.
const PortugueseDiacriticsReplaceID = "PT_DIACRITICS_REPLACE"

// PortugueseDiacriticsFilePath is the official resource path Java loads
// (PortugueseDiacriticsRule.FILE_NAME / getFileNames() → "/pt/diacritics.txt").
const PortugueseDiacriticsFilePath = "/pt/diacritics.txt"

// PortugueseDiacriticsURL ports PortugueseDiacriticsRule.getUrl()
// → Tools.getUrl("http://escreverbem.com.br/o-frances-no-portugues-2/").
const PortugueseDiacriticsURL = "http://escreverbem.com.br/o-frances-no-portugues-2/"

var (
	diacriticsOnce sync.Once
	diacriticsBase *rules.AbstractSimpleReplaceRule2
)

// loadDiacritics loads official diacritics.txt once (Java AbstractSimpleReplaceRule2 lazyInit).
func loadDiacritics() *rules.AbstractSimpleReplaceRule2 {
	diacriticsOnce.Do(func() {
		f, err := diacriticsFS.Open("data/diacritics.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of PortugueseDiacriticsRule + AbstractSimpleReplaceRule2 defaults:
		// - id PT_DIACRITICS_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " ou "
		// - CaseSensitivy.CI (base default)
		// - Locale pt via LanguageCode (getLocale → new Locale("pt"))
		// - useSubRuleSpecificIds()
		// - setDefaultOff()
		// - Categories.TYPOS, ITSIssueType.Misspelling
		// - URL escreverbem.com.br French-in-Portuguese
		// - example coupe → coupé
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   PortugueseDiacriticsReplaceID,
			Description:          "Palavras estrangeiras com diacríticos: $match",
			ShortMsg:             "A palavra estrangeira original tem diacrítico",
			MessageTemplate:      "'$match' é uma expressão estrangeira importada cuja grafia tem diacríticos. É preferível escrever $suggestions",
			SuggestionsSeparator: " ou ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "pt",
			SubRuleSpecificIDs:   true,
			Category:             rules.CatTypos.GetCategory(nil),
			IssueType:            rules.ITSMisspelling,
			DefaultOff:           true,
			URL:                  PortugueseDiacriticsURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, PortugueseDiacriticsFilePath); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("<marker>coupe</marker>"),
			rules.Fixed("<marker>coupé</marker>"),
		)
		diacriticsBase = base
	})
	return diacriticsBase
}

// PortugueseDiacriticsRule ports org.languagetool.rules.pt.PortugueseDiacriticsRule
// (extends AbstractSimpleReplaceRule2).
//
// Note: Java PortugueseDiacriticsRuleTest targets grammar rulegroup DIACRITICS
// (dialect bebé/bebê via full JLanguageTool), not this ASR2 class. Unit tests
// exercise PortugueseDiacriticsRule / official /pt/diacritics.txt directly
// (same bar as EN EnglishDiacriticsRule ACCEPT).
type PortugueseDiacriticsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewPortugueseDiacriticsRule ports PortugueseDiacriticsRule(ResourceBundle messages).
// Official data is go:embed of the same diacritics.txt Java loads from the classpath.
func NewPortugueseDiacriticsRule(messages map[string]string) *PortugueseDiacriticsRule {
	base := loadDiacritics()
	r := *base
	r.Messages = messages
	r.Category = rules.CatTypos.GetCategory(messages)
	r.IssueType = rules.ITSMisspelling
	r.DefaultOff = true
	r.URL = PortugueseDiacriticsURL
	return &PortugueseDiacriticsRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports PortugueseDiacriticsRule.getFileNames() → singleton "/pt/diacritics.txt".
func (r *PortugueseDiacriticsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{PortugueseDiacriticsFilePath}
}

// GetShort ports PortugueseDiacriticsRule.getShort().
func (r *PortugueseDiacriticsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseDiacriticsRule.getMessage().
func (r *PortugueseDiacriticsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports PortugueseDiacriticsRule.getSuggestionsSeparator() → " ou ".
func (r *PortugueseDiacriticsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports PortugueseDiacriticsRule.getLocale() → Locale("pt").
func (r *PortugueseDiacriticsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from diacritics.txt).
func (r *PortugueseDiacriticsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
