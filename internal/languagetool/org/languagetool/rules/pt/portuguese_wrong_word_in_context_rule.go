package pt

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

// PortugueseWrongWordInContextRule ports org.languagetool.rules.pt.PortugueseWrongWordInContextRule.
type PortugueseWrongWordInContextRule struct {
	*rules.WrongWordInContextRule
}

func NewPortugueseWrongWordInContextRule(messages map[string]string) *PortugueseWrongWordInContextRule {
	base := &rules.WrongWordInContextRule{
		Messages:           messages,
		ID:                 "PORTUGUESE_WRONG_WORD_IN_CONTEXT",
		Description:        "Confusão de palavra dentro do contexto (p.ex. infligir/infringir, etc.)",
		MessageString:      "Pretende dizer <suggestion>$SUGGESTION</suggestion> em vez de $WRONGWORD?",
		ShortMessageString: "Possível confusão de termos. Verifique.",
		LongMessageString:  "Considere <suggestion>$SUGGESTION</suggestion>, i.e. $EXPLANATION_SUGGESTION, em vez de '$WRONGWORD', i.e. $EXPLANATION_WRONGWORD?",
		LanguageCode:       "pt",
		Entries:            loadWrongWordInContext(),
	}
	// Java: getCategoryString → "Confusão de Palavras: $match"
	rules.InitWrongWordInContextMeta(base, messages, "Confusão de Palavras: $match")
	return &PortugueseWrongWordInContextRule{WrongWordInContextRule: base}
}

func (r *PortugueseWrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WrongWordInContextRule.Match(sentence)
}
