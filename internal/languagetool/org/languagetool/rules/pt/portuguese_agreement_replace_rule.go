package pt

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/AOreplace.txt
var aoReplaceFS embed.FS

// PortugueseAgreementReplaceID ports
// PortugueseAgreementReplaceRule.PORTUGUESE_AGREEMENT_REPLACE_RULE.
const PortugueseAgreementReplaceID = "PT_AGREEMENT_REPLACE"

// PortugueseAgreementReplaceFilePath is the official classpath resource Java loads
// (static loadFromPath("/pt/AOreplace.txt")).
const PortugueseAgreementReplaceFilePath = "/pt/AOreplace.txt"

// PortugueseAgreementReplaceURL ports PortugueseAgreementReplaceRule.getUrl()
// → Tools.getUrl("https://pt.wikipedia.org/wiki/Lista_das_alterações_previstas_pelo_acordo_ortográfico_de_1990").
const PortugueseAgreementReplaceURL = "https://pt.wikipedia.org/wiki/Lista_das_alterações_previstas_pelo_acordo_ortográfico_de_1990"

var (
	aoReplaceOnce sync.Once
	aoReplaceMap  map[string][]string
)

// loadAOReplace ports static wrongWords = loadFromPath("/pt/AOreplace.txt").
func loadAOReplace() map[string][]string {
	aoReplaceOnce.Do(func() {
		f, err := aoReplaceFS.Open("data/AOreplace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		aoReplaceMap = m
	})
	return aoReplaceMap
}

// PortugueseAgreementReplaceRule ports
// org.languagetool.rules.pt.PortugueseAgreementReplaceRule
// (extends AbstractSimpleReplaceRule).
//
// Java setDefaultOff() is commented out — rule remains default ON (ASR base default).
// Wired from PortugalPortuguese (not base Brazilian Portuguese).
type PortugueseAgreementReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewPortugueseAgreementReplaceRule ports
// PortugueseAgreementReplaceRule(ResourceBundle messages, Language language).
// language is Portuguese (short code "pt") for Locale("pt") / StringTools.toId.
func NewPortugueseAgreementReplaceRule(messages map[string]string) *PortugueseAgreementReplaceRule {
	// Field twin of PortugueseAgreementReplaceRule + AbstractSimpleReplaceRule defaults:
	// - id PT_AGREEMENT_REPLACE
	// - static wrongWords from /pt/AOreplace.txt
	// - isCaseSensitive() false
	// - checkLemmas default true (not setCheckLemmas(false))
	// - getLocale → Locale("pt") via LanguageCode
	// - useSubRuleSpecificIds()
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - URL wikipedia AO90 list
	// - description / short / message hardcoded PT strings (not ResourceBundle keys)
	// - example abstracto → abstrato
	// - setDefaultOff commented in Java → default ON
	base := &rules.AbstractSimpleReplaceRule{
		Messages:           messages,
		WrongWords:         loadAOReplace(),
		CaseSensitive:      false,
		CheckLemmas:        true, // Java default; not setCheckLemmas(false)
		ID:                 PortugueseAgreementReplaceID,
		LanguageCode:       "pt",
		SubRuleSpecificIDs: true,
		Description:        "Palavras alteradas pelo Acordo Ortográfico de 90",
		ShortMsg:           "Forma do Acordo Ortográfico de 45.",
		Category:           rules.CatTypos.GetCategory(messages),
		IssueType:          rules.ITSMisspelling,
		URL:                PortugueseAgreementReplaceURL,
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage: "'" + tokenStr + "' é uma forma do antigo acordo ortográfico. … " + String.join(", ", replacements) + "."
			return "'" + tokenStr + "' é uma forma do antigo acordo ortográfico. No novo acordo ortográfico, a palavra escreve-se assim: " +
				strings.Join(replacements, ", ") + "."
		},
	}
	// Java: abstracto → abstrato
	base.AddExamplePair(
		rules.Wrong("<marker>abstracto</marker>"),
		rules.Fixed("<marker>abstrato</marker>"),
	)
	return &PortugueseAgreementReplaceRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports PortugueseAgreementReplaceRule.getShort().
func (r *PortugueseAgreementReplaceRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports PortugueseAgreementReplaceRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real token and suggestions.
func (r *PortugueseAgreementReplaceRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports PortugueseAgreementReplaceRule.getLocale() → Locale("pt").
func (r *PortugueseAgreementReplaceRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports PortugueseAgreementReplaceRule.getWrongWords() → static map.
func (r *PortugueseAgreementReplaceRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *PortugueseAgreementReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}

