package nl

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

// DutchWrongWordInContextRule ports org.languagetool.rules.nl.DutchWrongWordInContextRule.
type DutchWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

func NewDutchWrongWordInContextRule(messages map[string]string) *DutchWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages:           messages,
		ID:                 "DUTCH_WRONG_WORD_IN_CONTEXT",
		Description:        "Woordverwarring: $match",
		MessageString:      "Mogelijk verwarring: Bedoelde u <suggestion>$SUGGESTION</suggestion> i.p.v. '$WRONGWORD'?",
		ShortMessageString: "Mogelijk verwarring",
		LongMessageString:  "Mogelijk verwarring: Bedoelde u <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) i.p.v. '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?",
		LanguageCode:       "nl",
		Entries:            loadWrongWordInContext(),
	}
	return &DutchWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *DutchWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
