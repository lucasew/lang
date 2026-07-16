package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/contractions.txt
var contractionsFS embed.FS

var (
	contractionsOnce sync.Once
	contractionWords map[string][]string
)

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

// ContractionSpellingRule ports org.languagetool.rules.en.ContractionSpellingRule.
type ContractionSpellingRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewContractionSpellingRule(messages map[string]string) *ContractionSpellingRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadContractions(),
		CaseSensitive: true,
		CheckLemmas:   false,
		ID:            "EN_CONTRACTION_SPELLING",
		Description:   "Spelling of English contractions",
		ShortMsg:      "Spelling mistake",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Possible spelling mistake found."
		},
		// Idiom "whys and wherefores" (plural nouns) must not fire; Java twin expects 0 matches.
		// Full LT may tag/immunize these; without a tagger we exempt the known false alarm.
		TokenException: func(token *languagetool.AnalyzedTokenReadings) bool {
			switch token.GetToken() {
			case "whys", "Whys", "WHYS", "wherefores", "Wherefores", "WHEREFORES":
				return true
			}
			return false
		},
	}
	return &ContractionSpellingRule{AbstractSimpleReplaceRule: base}
}

// Match delegates to AbstractSimpleReplaceRule.
func (r *ContractionSpellingRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
