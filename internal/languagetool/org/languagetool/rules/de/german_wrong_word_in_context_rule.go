package de

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

// GermanWrongWordInContextRule ports org.languagetool.rules.de.GermanWrongWordInContextRule.
type GermanWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

func NewGermanWrongWordInContextRule(messages map[string]string) *GermanWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages:           messages,
		ID:                 "GERMAN_WRONG_WORD_IN_CONTEXT",
		Description:        "Mögliche Wortverwechslungen: $match",
		MessageString:      "Mögliche Wortverwechslung: Meinten Sie <suggestion>$SUGGESTION</suggestion> anstatt '$WRONGWORD'?",
		ShortMessageString: "Mögliche Wortverwechslung",
		LongMessageString:  "Mögliche Wortverwechslung: Meinten Sie <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) anstatt '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?",
		LanguageCode:       "de",
		Entries:            loadWrongWordInContext(),
	}
	return &GermanWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *GermanWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
