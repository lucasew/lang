package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/english_contractions.txt
var engContrFS embed.FS

// EnglishContractionSpellingRuleID ports
// EnglishContractionSpellingRule.ENGLISH_CONTRACTION_SPELLING_RULE.
const EnglishContractionSpellingRuleID = "PT_ENGLISH_CONTRACTION_ORTHOGRAPHY"

// EnglishContractionSpellingFilePath is the official classpath resource Java loads
// (static loadFromPath("/pt/english_contractions.txt")).
const EnglishContractionSpellingFilePath = "/pt/english_contractions.txt"

var (
	engContrOnce sync.Once
	engContrMap  map[string][]string
)

// loadEnglishContractions ports static wrongWords = loadFromPath("/pt/english_contractions.txt").
func loadEnglishContractions() map[string][]string {
	engContrOnce.Do(func() {
		f, err := engContrFS.Open("data/english_contractions.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		engContrMap = m
	})
	return engContrMap
}

// EnglishContractionSpellingRule ports
// org.languagetool.rules.pt.EnglishContractionSpellingRule
// (extends AbstractSimpleReplaceRule).
//
// Copied into Portuguese from English to help code-switching and _english_ignore_
// tagging. Java setUrl(grammar-contractions) is COMMENTED OUT — no URL.
// No isTokenException override.
type EnglishContractionSpellingRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewEnglishContractionSpellingRule ports
// EnglishContractionSpellingRule(ResourceBundle messages, Language language).
// language is Portuguese (short code "pt") for Locale("pt").
func NewEnglishContractionSpellingRule(messages map[string]string) *EnglishContractionSpellingRule {
	// Field twin of EnglishContractionSpellingRule + AbstractSimpleReplaceRule:
	// - id PT_ENGLISH_CONTRACTION_ORTHOGRAPHY
	// - static wrongWords from /pt/english_contractions.txt
	// - isCaseSensitive() true
	// - setCheckLemmas(false)
	// - getLocale → Locale("pt") via LanguageCode
	// - Categories.TYPOS, ITSIssueType.Misspelling
	// - setUrl COMMENTED OUT in Java → URL empty / unset (do not invent EN URL)
	// - description / short / message hardcoded PT strings
	// - example pair whats → what's in Portuguese sentence
	// - no isTokenException
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadEnglishContractions(),
		CaseSensitive: true,  // Java isCaseSensitive()
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            EnglishContractionSpellingRuleID,
		LanguageCode:  "pt", // Java getLocale() → Locale("pt")
		Description:   "Ortografia de contrações inglesas",
		ShortMsg:      "Erro de ortografia inglesa",
		Category:      rules.CatTypos.GetCategory(messages),
		IssueType:     rules.ITSMisspelling,
		// URL deliberately empty: Java // setUrl(Tools.getUrl("https://languagetool.org/insights/post/grammar-contractions/"));
		MessageFn: func(tokenStr string, replacements []string) string {
			// Java getMessage: "Caso seja uma contração da língua inglesa, prefira \""
			//   + replacements.get(0) + "\"."
			pref := ""
			if len(replacements) > 0 {
				pref = replacements[0]
			}
			return "Caso seja uma contração da língua inglesa, prefira \"" + pref + "\"."
		},
	}
	// Java: addExamplePair(whats → what's) in Portuguese sentence
	base.AddExamplePair(
		rules.Wrong("Ele adorava assistir <marker>whats</marker> cooking às sextas-feiras."),
		rules.Fixed("Ele adorava assistir <marker>what's</marker> cooking às sextas-feiras."),
	)
	return &EnglishContractionSpellingRule{AbstractSimpleReplaceRule: base}
}

// GetShort ports EnglishContractionSpellingRule.getShort().
func (r *EnglishContractionSpellingRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports EnglishContractionSpellingRule.getMessage(token, replacements)
// with empty token/replacements for the static template surface (tests / metadata).
// Match-time messages go through MessageFn with real token and suggestions.
func (r *EnglishContractionSpellingRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	if r.MessageFn != nil {
		return r.MessageFn("", nil)
	}
	return ""
}

// GetLocaleCode ports EnglishContractionSpellingRule.getLocale() → Locale("pt").
func (r *EnglishContractionSpellingRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return ""
	}
	return r.LanguageCode
}

// GetWrongWords ports EnglishContractionSpellingRule.getWrongWords() → static map.
func (r *EnglishContractionSpellingRule) GetWrongWords() map[string][]string {
	if r == nil || r.AbstractSimpleReplaceRule == nil {
		return nil
	}
	return r.WrongWords
}

// Match ports AbstractSimpleReplaceRule.match.
func (r *EnglishContractionSpellingRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
