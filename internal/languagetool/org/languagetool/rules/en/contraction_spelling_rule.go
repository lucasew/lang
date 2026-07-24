package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/contractions.txt
var contractionsFS embed.FS

// ContractionSpellingRuleID ports ContractionSpellingRule.CONTRACTION_SPELLING_RULE.
const ContractionSpellingRuleID = "EN_CONTRACTION_SPELLING"

var (
	contractionsOnce sync.Once
	contractionWords map[string][]string
)

// loadContractions ports static loadFromPath("/en/contractions.txt") (official LT resource).
func loadContractions() map[string][]string {
	contractionsOnce.Do(func() {
		f, err := contractionsFS.Open("data/contractions.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		contractionWords = m
	})
	return contractionWords
}

// ContractionSpellingRule ports org.languagetool.rules.en.ContractionSpellingRule
// (extends AbstractSimpleReplaceRule). Case-sensitive; checkLemmas=false.
// No isTokenException in Java — "whys and wherefores" is handled by disambiguation
// IGNORE_SPELLING (IsIgnoredBySpeller), not surface invent.
type ContractionSpellingRule struct {
	*rules.AbstractSimpleReplaceRule
}

// NewContractionSpellingRule ports ContractionSpellingRule(ResourceBundle, Language).
// Language is only used by AbstractSimpleReplaceRule for sub-rule IDs (unused here).
func NewContractionSpellingRule(messages map[string]string) *ContractionSpellingRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadContractions(),
		CaseSensitive: true,  // Java isCaseSensitive()
		CheckLemmas:   false, // Java setCheckLemmas(false)
		ID:            ContractionSpellingRuleID,
		Description:   "Spelling of English contractions",
		ShortMsg:      "Spelling mistake", // Java getShort()
		// Java: Categories.TYPOS, Misspelling, grammar-contractions insights URL
		Category:  rules.CatTypos.GetCategory(messages),
		IssueType: rules.ITSMisspelling,
		URL:       "https://languagetool.org/insights/post/grammar-contractions/",
		// Java getMessage: always this string (locale en unused — case-sensitive).
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Possible spelling mistake found."
		},
	}
	// Java: addExamplePair(havent → haven't)
	base.AddExamplePair(
		rules.Wrong("We <marker>havent</marker> earned anything."),
		rules.Fixed("We <marker>haven't</marker> earned anything."),
	)
	return &ContractionSpellingRule{AbstractSimpleReplaceRule: base}
}

// Match ports AbstractSimpleReplaceRule.match (surface wrong-words only).
func (r *ContractionSpellingRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
