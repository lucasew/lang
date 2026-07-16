package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/coherency.txt
var coherencyFS embed.FS

var (
	coherencyOnce sync.Once
	coherencyData *rules.WordCoherencyData
)

func loadCoherencyData() *rules.WordCoherencyData {
	coherencyOnce.Do(func() {
		f, err := coherencyFS.Open("data/coherency.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := rules.LoadWordCoherencyData(f, "/pt/coherency.txt", false)
		if err != nil {
			panic(err)
		}
		coherencyData = d
	})
	return coherencyData
}

// PortugueseWordCoherencyRule ports org.languagetool.rules.pt.PortugueseWordCoherencyRule.
type PortugueseWordCoherencyRule struct {
	*rules.AbstractWordCoherencyRule
}

func NewPortugueseWordCoherencyRule(messages map[string]string) *PortugueseWordCoherencyRule {
	d := loadCoherencyData()
	base := &rules.AbstractWordCoherencyRule{
		Messages:    messages,
		ID:          "PT_WORD_COHERENCY",
		Description: "Consistência de palavras com grafias múltiplas",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		MessageFn: func(word1, word2 string) string {
			return "Não deve utilizar formas distintas de palavras com dupla grafia no mesmo texto. Escolha entre '" + word1 + "' e '" + word2 + "'."
		},
	}
	return &PortugueseWordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *PortugueseWordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
