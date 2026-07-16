package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wrongWordInContext.txt
var wrongWordInContextFS embed.FS

var (
	wwicOnce    sync.Once
	wwicEntries []rules.ContextWords
)

func loadWrongWordInContext() []rules.ContextWords {
	wwicOnce.Do(func() {
		f, err := wrongWordInContextFS.Open("data/wrongWordInContext.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		entries, err := rules.LoadWrongWordInContext(f)
		if err != nil {
			panic(err)
		}
		wwicEntries = entries
	})
	return wwicEntries
}

// CatalanWrongWordInContextRule ports org.languagetool.rules.ca.CatalanWrongWordInContextRule.
type CatalanWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

func NewCatalanWrongWordInContextRule(messages map[string]string) *CatalanWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages:           messages,
		ID:                 "CATALAN_WRONG_WORD_IN_CONTEXT",
		Description:        "Confusió segons el context: $match",
		MessageString:      "¿Volíeu dir <suggestion>$SUGGESTION</suggestion> en lloc de \"$WRONGWORD\"?",
		ShortMessageString: "Possible confusió",
		LongMessageString:  "¿Volíeu dir <suggestion>$SUGGESTION</suggestion> ($EXPLANATION_SUGGESTION) en lloc de \"$WRONGWORD\" ($EXPLANATION_WRONGWORD)?",
		LanguageCode:       "ca",
		MatchLemmas:        true,
		Entries:            loadWrongWordInContext(),
	}
	return &CatalanWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *CatalanWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
