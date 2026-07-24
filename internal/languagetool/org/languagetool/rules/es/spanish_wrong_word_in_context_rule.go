package es

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

// SpanishWrongWordInContextRule ports org.languagetool.rules.es.SpanishWrongWordInContextRule.
type SpanishWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

func NewSpanishWrongWordInContextRule(messages map[string]string) *SpanishWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages:           messages,
		ID:                 "SPANISH_WRONG_WORD_IN_CONTEXT",
		Description:        "Confusión según el contexto: $match",
		MessageString:      "¿Quería decir <suggestion>$SUGGESTION</suggestion> en vez de \"$WRONGWORD\"?",
		ShortMessageString: "Posible confusión",
		LongMessageString:  "¿Quería decir <suggestion>$SUGGESTION</suggestion> ($EXPLANATION_SUGGESTION) en vez de \"$WRONGWORD\" ($EXPLANATION_WRONGWORD)?",
		LanguageCode:       "es",
		MatchLemmas:        true,
		Entries:            loadWrongWordInContext(),
	}
	// Java: getCategoryString → "Confusiones"
	rules.InitWrongWordInContextMeta(base, messages, "Confusiones")
	return &SpanishWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *SpanishWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
