package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace-prestandard.txt
var prestandardFS embed.FS

var (
	prestandardOnce sync.Once
	prestandardMap  map[string][]string
)

func loadPrestandard() map[string][]string {
	prestandardOnce.Do(func() {
		f, err := prestandardFS.Open("data/replace-prestandard.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		prestandardMap = m
	})
	return prestandardMap
}

// PrestandardReplaceRule ports org.languagetool.rules.ga.PrestandardReplaceRule.
type PrestandardReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewPrestandardReplaceRule(messages map[string]string) *PrestandardReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadPrestandard(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_PRESTANDARD_REPLACE",
		Description:   "Litriú réamhchaighdeánach, m.sh., \"baoghal\" in áit \"baol\"",
		ShortMsg:      "Litriú réamhchaighdeánach",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Litriú réamhchaighdeánach:\"" + joinCommaGA(replacements) + "\"."
		},
	}
	// Java: baoghal → baol
	base.AddExamplePair(
		rules.Wrong("“Ní <marker>baoghal</marker> daoibh,” ar sise."),
		rules.Fixed("“Ní <marker>baol</marker> daoibh,” ar sise."),
	)
	return &PrestandardReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *PrestandardReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
