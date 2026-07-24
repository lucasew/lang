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
		ID:          "PT_WORD_COHERENCY",
		Description: "Consistência de palavras com grafias múltiplas",
		WordMap:     d.WordMap,
		ToBase:      d.ToBase,
		// Java: setCategory(STYLE); setLocQualityIssueType(Inconsistency)
		Category:  rules.CatStyle.GetCategory(messages),
		IssueType: rules.ITSInconsistency,
		MessageFn: func(word1, word2 string) string {
			return "Não deve utilizar formas distintas de palavras com dupla grafia no mesmo texto. Escolha entre '" + word1 + "' e '" + word2 + "'."
		},
	}
	rules.InitWordCoherencyMeta(base, messages)
	// Java: duradoiro → duradouro
	base.AddExamplePair(
		rules.Wrong("Foi um período duradouro. Tão marcante e <marker>duradoiro</marker> dificilmente será esquecido."),
		rules.Fixed("Foi um período duradouro. Tão marcante e <marker>duradouro</marker> dificilmente será esquecido."),
	)
	return &PortugueseWordCoherencyRule{AbstractWordCoherencyRule: base}
}

func (r *PortugueseWordCoherencyRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractWordCoherencyRule.Match(sentences)
}
